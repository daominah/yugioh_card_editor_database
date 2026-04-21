#!/usr/bin/env python3
"""
main.py: Color the fallback.svg skeleton using frame texture PNGs.

See fill_by_region_plan.md for the full plan. Steps:
  1. Start   : blank transparent canvas.
  2. Skeleton: paint SVG black strokes as opaque black onto the canvas.
  3. Copy    : for pixels in regions 1/2/3 that are not blackish, copy from source PNG.
               Pixels near SVG lines or region boundaries use a stricter luminance threshold.
               Save intermediate to removed_black/.
  4. Outliers: erase copied pixels whose hue deviates too far from the region average.
  5. Fill    : fill remaining transparent gaps using weighted random patch sampling.
  6. Border  : composite with border_strip.png and ATK/DEF strip.
  7. Skeleton: redraw SVG skeleton strokes on top so lines are crisp on the final output.

Regions (matching fill_by_region_plan.md):
  1 = between Inner card boundary and Art box  (yellowish)
  2 = between Effect box outer and Effect box inner  (same tone as region 1)
  3 = inside Effect box inner / text area  (lighter, yellowish-white)

Usage:
    python main.py                       # runs all entries in INPUT_CONFIGS
    python main.py source.png output.png # single file, default config
"""

import os
import sys
import time

import numpy as np
from PIL import Image

from border import add_border_strip
from copy_pixels import copy_pixels
from fill_gaps import fill_gaps
from svg_masks import create_region_mask, rasterize_skeleton, _level_badge_mask


_DIR = os.path.dirname(__file__)

# Final output; used as card frames for the Card Editor (Rush Duel layout).
OUTPUT_DIR = os.path.join(_DIR, "..", "web", "card_frame_rushduel")

# The same card frame skeleton in two formats:
# SVG_PATH is used for region geometry (create_region_mask, edge distance masks).
# FALLBACK_PNG is the SVG pre-rendered to PNG, used for the skeleton pixel mask
# (pixel-accurate stroke positions without PIL's stroke-centering approximation).
SVG_PATH = os.path.join(OUTPUT_DIR, "fallback.svg")
FALLBACK_PNG = os.path.join(OUTPUT_DIR, "fallback.png")

ATK_DEF_STRIP_SVG = os.path.join(OUTPUT_DIR, "atk_def_strip.svg")
BORDER_STRIP_PNG = os.path.join(OUTPUT_DIR, "border_strip.png")
INPUT_DIR = os.path.join(_DIR, "input_png")
REMOVED_BLACK_DIR = os.path.join(_DIR, "removed_black")

# ---------------------------------------------------------------------------
# Per-file config
# ---------------------------------------------------------------------------
# Each entry maps an input basename to a dict of parameter overrides.
# Keys match the defaults in _default_config() below.
# Add a new entry for each card type that needs different tuning.

INPUT_CONFIGS = {
    # "spell.png": {
    #     "COPY_LEVEL_BADGE": False,
    #     "ADD_ATK_DEF_STRIP": False,
    #     "LUMINANCE_THRESHOLD_NEAR_EDGE": 150,
    # },
    # "trap.png": {
    #     "COPY_LEVEL_BADGE": False,
    #     "ADD_ATK_DEF_STRIP": False,
    # },
    # "monster_normal.png": {
    #     "LUMINANCE_THRESHOLD_NEAR_EDGE": 180,
    # },
    # "monster_effect.png": {},
    # "monster_fusion.png": {},
    # "monster_ritual.png": {
    #     "COPY_ALL_IN_REGION": True,
    #     "ALPHA_THRESHOLD": 240,
    #     "SAMPLE_EXCLUDE_EDGE_PX": 8,
    # },
    # "monster_link.png": {},
    # "monster_synchro.png": {},
    # "monster_xyz.png": {  # input is very dark; skip luminance filter
    #     "COPY_ALL_IN_REGION": True,
    #     "ALPHA_THRESHOLD": 240,
    #     "SAMPLE_EXCLUDE_EDGE_PX": 8,
    # },
    # "monster_synchro.png": {
    #     "COPY_ALL_IN_REGION": True,
    #     "ALPHA_THRESHOLD": 240,
    #     "SAMPLE_EXCLUDE_EDGE_PX": 8,
    # },
    "monster_link.png": {
        "COPY_ALL_IN_REGION": True,
        "ALPHA_THRESHOLD": 240,
        "SAMPLE_EXCLUDE_EDGE_PX": 8,
    },
}


def _default_config():
    return {
        # Pixels with luminance below this are considered "blackish" and are not copied.
        # Luminance = 0.299R + 0.587G + 0.114B (range 0-255).
        "LUMINANCE_THRESHOLD": 80,

        # Pixels within EDGE_DISTANCE_PX of any SVG line use this stricter threshold,
        # since upscaling bleeds black stroke color further into the frame near edges.
        "LUMINANCE_THRESHOLD_NEAR_EDGE": 160,

        # How many pixels away from an SVG line counts as "near edge".
        "EDGE_DISTANCE_PX": 8,

        # The inner card boundary borders regions that should be fully filled with color,
        # so it uses a smaller near-edge radius than other SVG lines.
        "INNER_BOUNDARY_EDGE_DISTANCE_PX": 1,

        # The art box top and bottom horizontal edges accumulate more upscaling bleed
        # than the sides, so they get their own (usually larger) near-edge radius.
        "ART_BOX_TB_EDGE_DISTANCE_PX": 24,

        # Minimum luminance a pixel must have to be copied into REGION_3.
        # Filters out darker bleed-in from upscaling misalignment at the border.
        "REGION_3_MIN_LUMINANCE": 216,

        # Pixels inside REGION_3 but within this many pixels of the region boundary
        # also use LUMINANCE_THRESHOLD_NEAR_EDGE, because the source image has misaligned
        # dark bleed near the effect box inner edge that the skeleton distance alone misses.
        "REGION_3_EDGE_DISTANCE_PX": 20,

        # Step 4: remove copied pixels whose hue deviates too far from the region average.
        # Catches off-tone bleed that the blackish check cannot catch.
        # Hue is on a 0-360 color wheel:
        # 0/360=red, 60=yellow, 120=green, 180=cyan, 240=blue, 300=magenta.
        # Set OUTLIER_MAX_HUE_DEV to 0 to disable this filter.
        "OUTLIER_MAX_HUE_DEV": 30,  # degrees (0 = disabled)

        # Whether to draw the level badge circle stroke in the skeleton.
        # False = no ring drawn; the badge area stays fully transparent in the output.
        "COPY_LEVEL_BADGE": True,

        # Pixels with alpha below this are treated as transparent.
        "ALPHA_THRESHOLD": 32,

        # Number of nearest already-filled pixels to sample from when filling gaps.
        "RANDOM_FILL_K": 64,
        # Fixed seed so the mottled texture pattern is deterministic across runs.
        # Change to get a different random arrangement of the texture.
        "RANDOM_FILL_SEED": 1,

        # Whether to add the semi-transparent dark strip behind ATK/DEF stats.
        # Spell/Trap cards do not have ATK/DEF, so they should set this to False.
        "ADD_ATK_DEF_STRIP": True,

        # When True, copy ALL opaque region pixels from the source in Step 3
        # without any luminance/blackish filtering, then skip Step 4 (hue outliers).
        # Use for inputs with a very dark or blackish tone where the normal
        # luminance threshold would discard most pixels.
        "COPY_ALL_IN_REGION": False,

        # Pixels within this many pixels of the skeleton are excluded from the gap-fill
        # sample pool. Stroke bleed from upscaling darkens these pixels; excluding them
        # from sampling prevents dark corners when gaps are near the skeleton boundary.
        # The pixels themselves remain in the canvas unchanged. 0 = disabled.
        "SAMPLE_EXCLUDE_EDGE_PX": 0,
    }


def _config_for(basename):
    cfg = _default_config()
    cfg.update(INPUT_CONFIGS.get(basename, {}))
    return cfg


# ---------------------------------------------------------------------------
# Main pipeline
# ---------------------------------------------------------------------------

def run(source_path, output_path):
    basename = os.path.basename(source_path)
    cfg = _config_for(basename)

    src = Image.open(source_path).convert("RGBA")
    arr = np.array(src, dtype=np.uint8)
    h, w = arr.shape[:2]

    # Step 1: blank canvas
    canvas = np.zeros((h, w, 4), dtype=np.uint8)
    print(f"Canvas: {w}x{h}  config: {basename}")

    # Step 2: draw skeleton
    skeleton = rasterize_skeleton(FALLBACK_PNG, w, h, cfg)
    if not cfg["COPY_LEVEL_BADGE"]:
        skeleton &= ~_level_badge_mask(SVG_PATH, w, h)
    canvas[skeleton] = [0, 0, 0, 255]
    print(f"Skeleton: {skeleton.sum()} black pixels painted.")

    # Steps 3-4: copy pixels from source, filter blackish, remove hue outliers
    region_mask = create_region_mask(SVG_PATH, w, h)
    copy_pixels(canvas, arr, region_mask, skeleton, cfg, SVG_PATH, REMOVED_BLACK_DIR, basename)

    # Steps 5, 5.5, 5.6: fill gaps, blur seams, fill logo gradient
    fill_gaps(canvas, region_mask, skeleton, cfg)

    # Step 6: composite border strip and ATK/DEF strip
    filled_img = Image.fromarray(canvas, "RGBA")
    result = add_border_strip(filled_img, cfg, ATK_DEF_STRIP_SVG, BORDER_STRIP_PNG)

    # Step 7: redraw skeleton on final composite so strokes are crisp
    result_arr = np.array(result)
    result_arr[skeleton] = [0, 0, 0, 255]
    Image.fromarray(result_arr, "RGBA").save(output_path)
    print(f"Saved to {output_path}")


if __name__ == "__main__":
    begin_t = time.time()
    print(f"begin main at {time.strftime('%Y-%m-%dT%H:%M:%Sz')}")

    if len(sys.argv) == 3:
        run(sys.argv[1], sys.argv[2])
    else:
        for name in INPUT_CONFIGS:
            print("_" * 40)
            source = os.path.join(INPUT_DIR, name)
            output = os.path.join(OUTPUT_DIR, name)
            if os.path.exists(source):
                run(source, output)
            else:
                print(f"Skipping {name}: {source} not found.")

    print(f"end main. duration: {time.time() - begin_t:.2f} seconds.")
