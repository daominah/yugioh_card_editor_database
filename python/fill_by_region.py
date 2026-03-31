#!/usr/bin/env python3
"""
fill_by_region.py: Color the fallback.svg skeleton using frame texture PNGs.

See fill_by_region_plan.md for the full plan. Steps:
  1. Start   : blank transparent canvas.
  2. Skeleton: paint SVG black strokes as opaque black onto the canvas.
  3. Copy    : for pixels in regions 1/2/3 that are not blackish, copy from source PNG.
               Pixels near SVG lines or region boundaries use a stricter luminance threshold.
               Save intermediate to removed_black/.
  4. Outliers: erase copied pixels whose hue deviates too far from the region average.
  5. Fill    : fill remaining transparent gaps using weighted random patch sampling.
  6. Border  : composite with border_strip.png and save to web/card_frame_rushduel/.

Regions (matching fill_by_region_plan.md):
  1 = between Inner card boundary and Art box  (yellowish)
  2 = between Effect box outer and Effect box inner  (same tone as region 1)
  3 = inside Effect box inner / text area  (lighter, yellowish-white)

Usage:
    python fill_by_region.py                       # runs all entries in INPUT_CONFIGS
    python fill_by_region.py source.png output.png # single file, default config
"""

import io
import math
import os
import re
import sys
import time
import xml.etree.ElementTree as ET

import numpy as np
from PIL import Image, ImageDraw
from scipy.ndimage import binary_dilation, distance_transform_edt, gaussian_filter, uniform_filter
from scipy.spatial import cKDTree


_DIR = os.path.dirname(__file__)

# final output, will be used as card frames for the Card Editor (Rush Duel layout)
OUTPUT_DIR = os.path.join(_DIR, "..", "web", "card_frame_rushduel")

SVG_PATH = os.path.join(OUTPUT_DIR, "fallback.svg")
ATK_DEF_STRIP_SVG = os.path.join(OUTPUT_DIR, "atk_def_strip.svg")
BORDER_STRIP_PNG = os.path.join(OUTPUT_DIR, "border_strip.png")
INPUT_DIR = os.path.join(_DIR, "input_png")

# inte
REMOVED_BLACK_DIR = os.path.join(_DIR, "removed_black")

# ---------------------------------------------------------------------------
# Per-file config
# ---------------------------------------------------------------------------
# Each entry maps an input basename to a dict of parameter overrides.
# Keys match the defaults in _default_config() below.
# Add a new entry for each card type that needs different tuning.

INPUT_CONFIGS = {
    "spell.png": {
        "COPY_LEVEL_BADGE": False,
        "ADD_ATK_DEF_STRIP": False,
        "LUMINANCE_THRESHOLD_NEAR_EDGE": 150,
    },
    "trap.png": {
        "COPY_LEVEL_BADGE": False,
        "ADD_ATK_DEF_STRIP": False,
    },
    "monster_normal.png": {
        "LUMINANCE_THRESHOLD_NEAR_EDGE": 180,
    },
    "monster_effect.png": {

    },
    "monster_fusion.png": {
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

        # Step 3.5: remove copied pixels whose hue deviates too far from the region average.
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

    }


def _config_for(basename):
    cfg = _default_config()
    cfg.update(INPUT_CONFIGS.get(basename, {}))
    return cfg


# The 3 regions from fill_plan.md.
REGION_1 = 1  # between Inner card boundary and Art box (yellowish)
REGION_2 = 2  # between Effect box outer and Effect box inner (same tone as region 1)
REGION_3 = 3  # inside Effect box inner / text area (lighter, yellowish-white)


# ---------------------------------------------------------------------------
# SVG helpers
# ---------------------------------------------------------------------------

def _parse_polyline_points(points_str, sx, sy):
    pts = []
    for t in points_str.split():
        t = t.strip().rstrip(",")
        if "," in t:
            px, py = t.split(",", 1)
            pts.append((float(px) * sx, float(py) * sy))
    return pts


def _poly_mask(pts, width, height):
    img = Image.new("L", (width, height), 0)
    if len(pts) >= 3:
        ImageDraw.Draw(img).polygon(pts, fill=255)
    return np.array(img) > 0


def _rect_mask(x, y, w, h, sx, sy, width, height):
    img = Image.new("L", (width, height), 0)
    ImageDraw.Draw(img).rectangle(
        [(x * sx, y * sy), ((x + w) * sx, (y + h) * sy)], fill=255
    )
    return np.array(img) > 0


def _ellipse_mask(cx, cy, r, sx, sy, width, height):
    img = Image.new("L", (width, height), 0)
    ImageDraw.Draw(img).ellipse(
        [((cx - r) * sx, (cy - r) * sy), ((cx + r) * sx, (cy + r) * sy)], fill=255
    )
    return np.array(img) > 0


def _svg_scale(svg_path, width, height):
    tree = ET.parse(svg_path)
    root = tree.getroot()
    vb = root.get("viewBox", f"0 0 {width} {height}").split()
    return root, width / float(vb[2]), height / float(vb[3])


def _title_of(elem):
    for child in elem:
        if child.tag.split("}")[-1] == "title":
            return (child.text or "").strip().lower()
    return ""


def rasterize_skeleton(svg_path, width, height, cfg):
    """Return bool mask (H x W) where SVG black stroke/fill pixels are True."""
    root, sx, sy = _svg_scale(svg_path, width, height)
    canvas = Image.new("L", (width, height), 0)
    draw = ImageDraw.Draw(canvas)

    def sw(elem):
        return max(1, round(float(elem.get("stroke-width", "1")) * (sx + sy) / 2))

    def is_black(c):
        return c and c.lower() not in ("none", "")

    for elem in root.iter():
        tag = elem.tag.split("}")[-1]
        stroke = elem.get("stroke")
        fill = elem.get("fill", "none")
        w = sw(elem)

        if tag == "rect":
            x, y = float(elem.get("x", 0)), float(elem.get("y", 0))
            rw, rh = float(elem.get("width")), float(elem.get("height"))
            # PIL draws rectangle/ellipse strokes inside the bounding box.
            # SVG strokes are centered on the path (half outside, half inside).
            # Expand by w//2 so the stroke center aligns with the SVG path.
            hw = w // 2
            xy = [(x * sx - hw, y * sy - hw), ((x + rw) * sx + hw, (y + rh) * sy + hw)]
            if is_black(stroke):
                draw.rectangle(xy, outline=255, width=w)

        elif tag == "circle":
            if _title_of(elem) == "level badge" and not cfg["COPY_LEVEL_BADGE"]:
                continue
            cx, cy, r = float(elem.get("cx")), float(elem.get("cy")), float(elem.get("r"))
            # Same stroke-centering correction as rect above.
            hw = w // 2
            xy = [((cx - r) * sx - hw, (cy - r) * sy - hw), ((cx + r) * sx + hw, (cy + r) * sy + hw)]
            if is_black(stroke):
                draw.ellipse(xy, outline=255, width=w)

        elif tag == "polyline":
            pts = _parse_polyline_points(elem.get("points", ""), sx, sy)
            if len(pts) >= 2:
                if is_black(stroke):
                    draw.line(pts, fill=255, width=w)
                if fill.lower() == "black":
                    draw.polygon(pts, fill=255)

        elif tag == "path":
            if fill.lower() == "black":
                # Simple filled path (footer etc.)
                tokens = re.findall(r'[MHVZmhvz]|[-+]?[0-9]*\.?[0-9]+', elem.get("d", ""))
                pts, sub, cx2, cy2 = [], [], 0.0, 0.0
                i = 0
                while i < len(tokens):
                    cmd = tokens[i];
                    i += 1
                    if cmd == 'M':
                        if sub: pts.append(sub)
                        cx2, cy2 = float(tokens[i]), float(tokens[i + 1]);
                        i += 2
                        sub = [(cx2 * sx, cy2 * sy)]
                    elif cmd == 'H':
                        cx2 = float(tokens[i]);
                        i += 1;
                        sub.append((cx2 * sx, cy2 * sy))
                    elif cmd == 'V':
                        cy2 = float(tokens[i]);
                        i += 1;
                        sub.append((cx2 * sx, cy2 * sy))
                    elif cmd in ('Z', 'z'):
                        if sub: sub.append(sub[0]); pts.append(sub); sub = []
                for s in pts:
                    if len(s) >= 3:
                        draw.polygon(s, fill=255)

    return np.array(canvas) > 0


def create_region_mask(svg_path, width, height):
    """
    Return uint8 array (H x W) with values 1, 2, 3 for the 3 target regions.
    0 = not a target region. Later assignments override earlier ones.

    REGION_1: inner card body (inside inner boundary, minus art box and below).
    REGION_2: effect box border (inside outer rect, minus inner polyline).
    REGION_3: effect box text area (inside inner polyline).
    """
    root, sx, sy = _svg_scale(svg_path, width, height)
    mask = np.zeros((height, width), dtype=np.uint8)

    # Paint REGION_1 as the inner card body first; overrides follow.
    inner = _rect_mask(40, 40, 1100, 1640, sx, sy, width, height)
    mask[inner] = REGION_1

    for elem in root.iter():
        tag = elem.tag.split("}")[-1]
        t = _title_of(elem)

        if tag == "polyline":
            pts = _parse_polyline_points(elem.get("points", ""), sx, sy)
            m = _poly_mask(pts, width, height)
            if "art box" in t:
                mask[m] = 0  # art box: not a target region
            elif "effect box inner" in t:
                mask[m] = REGION_3  # innermost: text area
            elif "footer" in t:
                mask[m] = 0
            elif "bottom logo" in t:
                mask[m] = 0

        elif tag == "rect":
            x = float(elem.get("x", 0))
            y = float(elem.get("y", 0))
            w = float(elem.get("width"))
            h = float(elem.get("height"))
            if "effect box outer" in t:
                # Paint the outer rect as REGION_2; inner polyline overrides to REGION_3 above.
                mask[_rect_mask(x, y, w, h, sx, sy, width, height)] = REGION_2

        elif tag == "circle":
            cx, cy, r = float(elem.get("cx")), float(elem.get("cy")), float(elem.get("r"))
            mask[_ellipse_mask(cx, cy, r, sx, sy, width, height)] = 0  # circles: not target

    return mask


# ---------------------------------------------------------------------------
# Fill logic
# ---------------------------------------------------------------------------

def _inner_boundary_mask(svg_path, width, height):
    """Return bool mask (H x W) marking the Inner card boundary rect stroke."""
    root, sx, sy = _svg_scale(svg_path, width, height)
    canvas = Image.new("L", (width, height), 0)
    draw = ImageDraw.Draw(canvas)
    for elem in root.iter():
        tag = elem.tag.split("}")[-1]
        if tag != "rect":
            continue
        if _title_of(elem) != "inner card boundary":
            continue
        x = float(elem.get("x", 0))
        y = float(elem.get("y", 0))
        w = float(elem.get("width"))
        h = float(elem.get("height"))
        sw = max(1, round(float(elem.get("stroke-width", "1")) * (sx + sy) / 2))
        draw.rectangle([(x * sx, y * sy), ((x + w) * sx, (y + h) * sy)], outline=255, width=sw)
    return np.array(canvas) > 0


def _art_box_tb_mask(svg_path, width, height):
    """
    Return bool mask (H x W) marking only the top and bottom horizontal segments
    of the Art box polyline. Used for a separate near-edge distance radius.
    A segment is considered horizontal when its two endpoints share the same y value.
    """
    root, sx, sy = _svg_scale(svg_path, width, height)
    canvas = Image.new("L", (width, height), 0)
    draw = ImageDraw.Draw(canvas)
    for elem in root.iter():
        tag = elem.tag.split("}")[-1]
        if tag != "polyline":
            continue
        if _title_of(elem) != "art box":
            continue
        pts = _parse_polyline_points(elem.get("points", ""), sx, sy)
        sw = max(1, round(float(elem.get("stroke-width", "1")) * (sx + sy) / 2))
        for i in range(len(pts) - 1):
            x0, y0 = pts[i]
            x1, y1 = pts[i + 1]
            if abs(y0 - y1) < 1:  # horizontal segment
                draw.line([(x0, y0), (x1, y1)], fill=255, width=sw)
    return np.array(canvas) > 0


def _is_blackish(arr, skeleton, inner_boundary, art_box_tb, cfg):
    """
    Return bool mask (H x W): True where pixel is considered blackish.
    Pixels near SVG lines use LUMINANCE_THRESHOLD_NEAR_EDGE (stricter) because
    upscaling bleeds dark stroke color further into the frame area near edges.
    The inner card boundary uses a reduced 1px radius since it borders regions
    that should be fully filled with color (not near any bleed source).
    The art box top/bottom horizontal edges use their own radius ART_BOX_TB_EDGE_DISTANCE_PX.
    """
    r = arr[:, :, 0].astype(np.float32)
    g = arr[:, :, 1].astype(np.float32)
    b = arr[:, :, 2].astype(np.float32)
    lum = 0.299 * r + 0.587 * g + 0.114 * b
    # Exclude inner boundary from the main skeleton so it doesn't contribute EDGE_DISTANCE_PX.
    skeleton_no_inner = skeleton & ~inner_boundary
    dist_to_skeleton = distance_transform_edt(~skeleton_no_inner)
    dist_to_inner = distance_transform_edt(~inner_boundary)
    dist_to_art_tb = distance_transform_edt(~art_box_tb)
    near_edge = (
            (dist_to_skeleton <= cfg["EDGE_DISTANCE_PX"]) |
            (dist_to_inner <= cfg["INNER_BOUNDARY_EDGE_DISTANCE_PX"]) |
            (dist_to_art_tb <= cfg["ART_BOX_TB_EDGE_DISTANCE_PX"])
    )
    threshold = np.where(near_edge, cfg["LUMINANCE_THRESHOLD_NEAR_EDGE"], cfg["LUMINANCE_THRESHOLD"])
    return lum < threshold


def remove_hue_outliers(canvas, region_mask, cfg):
    """
    Step 3.5: erase copied pixels whose hue deviates more than OUTLIER_MAX_HUE_DEV
    degrees from the average hue of their region.
    Only pixels with saturation >= 20 (out of 255) are checked; low-saturation pixels
    have unstable hue and are left to the blackish filter instead.
    Erased pixels become transparent so the gap-fill step covers them.
    Disabled when OUTLIER_MAX_HUE_DEV == 0.
    """
    max_dev = cfg["OUTLIER_MAX_HUE_DEV"]
    if max_dev == 0:
        return

    min_sat = 20
    alpha_threshold = cfg["ALPHA_THRESHOLD"]

    r = canvas[:, :, 0].astype(np.float32)
    g = canvas[:, :, 1].astype(np.float32)
    b = canvas[:, :, 2].astype(np.float32)

    # HSV saturation (0-255): (max - min) / max * 255
    cmax = np.maximum(np.maximum(r, g), b)
    cmin = np.minimum(np.minimum(r, g), b)
    delta = cmax - cmin
    with np.errstate(invalid="ignore", divide="ignore"):
        sat = np.where(cmax > 0, delta / cmax * 255.0, 0.0)

    # Hue in degrees [0, 360)
    with np.errstate(invalid="ignore", divide="ignore"):
        hue = np.zeros_like(r)
        m = delta > 0
        mr = m & (cmax == r)
        mg = m & (cmax == g)
        mb = m & (cmax == b)
        hue[mr] = 60.0 * (((g[mr] - b[mr]) / delta[mr]) % 6)
        hue[mg] = 60.0 * (((b[mg] - r[mg]) / delta[mg]) + 2)
        hue[mb] = 60.0 * (((r[mb] - g[mb]) / delta[mb]) + 4)

    opaque = canvas[:, :, 3] >= alpha_threshold

    total_removed = 0
    for region_id in (REGION_1, REGION_2, REGION_3):
        in_region = region_mask == region_id
        checkable = in_region & opaque & (sat >= min_sat)
        if checkable.sum() == 0:
            continue
        avg_hue = float(hue[checkable].mean())
        diff = np.abs(hue - avg_hue)
        circ_diff = np.minimum(diff, 360.0 - diff)
        outlier = in_region & opaque & (sat >= min_sat) & (circ_diff > max_dev)
        canvas[outlier, 3] = 0
        total_removed += int(outlier.sum())
        print(f"  Region {region_id}: avg hue {avg_hue:.1f} deg, removed {outlier.sum()} hue outliers.")

    print(f"Hue outlier removal: {total_removed} pixels erased total.")


def _sample_gaps(canvas, gap_coords, filled_coords, cfg, rng):
    """Fill gap_coords by weighted random sampling from filled_coords.
    Returns nothing; writes directly into canvas."""
    k = cfg["RANDOM_FILL_K"]
    actual_k = min(k, len(filled_coords))
    tree = cKDTree(filled_coords)
    distances, indices = tree.query(gap_coords, k=actual_k)

    if actual_k == 1:
        distances = distances[:, np.newaxis]
        indices = indices[:, np.newaxis]

    # Power > 1 biases sampling toward nearer pixels; 1.5 is a middle ground between
    # moderate locality (1.0) and near-exclusive nearest-neighbor (2.0+).
    # weights = 1.0 / (distances + 1.0) ** 1.5
    weights = 1.0 / (distances + 1.0)
    weights /= weights.sum(axis=1, keepdims=True)

    cumw = np.cumsum(weights, axis=1)
    r = rng.random(len(gap_coords))
    chosen_col = (cumw < r[:, np.newaxis]).sum(axis=1).clip(0, actual_k - 1)

    src_idx = indices[np.arange(len(gap_coords)), chosen_col]
    src_yx = filled_coords[src_idx]

    canvas[gap_coords[:, 0], gap_coords[:, 1]] = canvas[src_yx[:, 0], src_yx[:, 1]]
    canvas[gap_coords[:, 0], gap_coords[:, 1], 3] = 255


def weighted_random_fill(canvas, region_mask, region_id, skeleton, cfg):
    """Fill transparent pixels in region using weighted random sampling from filled pixels.
    Skeleton pixels are excluded from the source pool.

    For REGION_2: each gap pixel checks how many filled REGION_2 pixels exist within a
    local 100x100 window. If >= 100, it samples from REGION_2 (spatially nearby, correct
    tone). If not, it falls back to use REGION_1 color."""

    k = cfg["RANDOM_FILL_K"]
    seed = cfg["RANDOM_FILL_SEED"]
    alpha_threshold = cfg["ALPHA_THRESHOLD"]

    rng = np.random.default_rng(seed)
    in_region = region_mask == region_id
    opaque = canvas[:, :, 3] >= alpha_threshold

    filled_own = in_region & opaque & ~skeleton
    gaps = in_region & ~opaque

    gap_coords = np.argwhere(gaps)
    if len(gap_coords) == 0:
        print(f"  Region {region_id}: no gaps to fill.")
        return

    if region_id != REGION_2:
        filled_coords = np.argwhere(filled_own)
        if len(filled_coords) == 0:
            print(f"  Region {region_id}: no source pixels, cannot fill {len(gap_coords)} gaps.")
            return
        _sample_gaps(canvas, gap_coords, filled_coords, cfg, rng)
        print(f"  Region {region_id}: filled {len(gap_coords)} gaps from {len(filled_coords)} source pixels.")
        return

    # REGION_2: per-gap local density check.
    # uniform_filter computes a sliding-window mean in O(W*H); multiply by window area
    # to get the local filled-pixel count for each position.
    window = 100
    min_filled = 100
    density = uniform_filter(filled_own.astype(np.float32), size=window) * window * window

    gap_density = density[gap_coords[:, 0], gap_coords[:, 1]]
    dense_mask = gap_density >= min_filled
    dense_gaps = gap_coords[dense_mask]
    sparse_gaps = gap_coords[~dense_mask]

    filled_own_coords = np.argwhere(filled_own)
    filled_r1_coords = np.argwhere((region_mask == REGION_1) & opaque & ~skeleton)

    if len(dense_gaps) > 0:
        if len(filled_own_coords) > 0:
            _sample_gaps(canvas, dense_gaps, filled_own_coords, cfg, rng)
        else:
            sparse_gaps = gap_coords  # no own pixels at all, treat everything as sparse

    if len(sparse_gaps) > 0:
        if len(filled_r1_coords) == 0:
            print(f"  Region {region_id}: no REGION_1 fallback pixels for {len(sparse_gaps)} sparse gaps.")
        else:
            _sample_gaps(canvas, sparse_gaps, filled_r1_coords, cfg, rng)

    print(f"  Region {region_id}: filled {len(gap_coords)} gaps "
          f"({len(dense_gaps)} from own, {len(sparse_gaps)} from REGION_1 fallback).")


# ---------------------------------------------------------------------------
# Composite preview
# ---------------------------------------------------------------------------

def add_border_strip(base_img, output_path, cfg):
    """Composite the filled image with the border strip and ATK/DEF dark
    strip overlays, then save to output_path."""
    w, h = base_img.size

    if cfg["ADD_ATK_DEF_STRIP"]:
        # Rasterize atk_def_strip.svg directly (single source of truth
        # for shape, color, and opacity).
        import cairosvg
        png_data = cairosvg.svg2png(url=ATK_DEF_STRIP_SVG,
                                    output_width=w, output_height=h)
        strip_layer = Image.open(io.BytesIO(png_data)).convert("RGBA")
        base_img = Image.alpha_composite(base_img, strip_layer)

    border = Image.open(BORDER_STRIP_PNG).convert("RGBA").resize(base_img.size, Image.LANCZOS)
    result = Image.alpha_composite(base_img, border)
    result.save(output_path)
    print(f"Saved to {output_path}")


# ---------------------------------------------------------------------------
# Main
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
    skeleton = rasterize_skeleton(SVG_PATH, w, h, cfg)
    canvas[skeleton] = [0, 0, 0, 255]
    print(f"Skeleton: {skeleton.sum()} black pixels painted.")

    # Step 3: copy good pixels per region
    region_mask = create_region_mask(SVG_PATH, w, h)
    inner_boundary = _inner_boundary_mask(SVG_PATH, w, h)
    art_box_tb = _art_box_tb_mask(SVG_PATH, w, h)
    blackish = _is_blackish(arr, skeleton, inner_boundary, art_box_tb, cfg)
    opaque = arr[:, :, 3] >= cfg["ALPHA_THRESHOLD"]

    r_chan = arr[:, :, 0].astype(np.float32)
    g_chan = arr[:, :, 1].astype(np.float32)
    b_chan = arr[:, :, 2].astype(np.float32)
    luminance = 0.299 * r_chan + 0.587 * g_chan + 0.114 * b_chan

    # Extra strict mask for pixels near the REGION_3 boundary.
    # The source image has dark bleed at the effect box inner edge due to upscaling
    # misalignment; these pixels are often >EDGE_DISTANCE_PX from the SVG skeleton,
    # so the skeleton-based near-edge threshold alone does not catch them.
    region3_bool = region_mask == REGION_3
    dist_inside_region3 = distance_transform_edt(region3_bool)
    near_region3_boundary = region3_bool & (dist_inside_region3 <= cfg["REGION_3_EDGE_DISTANCE_PX"])
    extra_blackish_region3 = near_region3_boundary & (luminance < cfg["LUMINANCE_THRESHOLD_NEAR_EDGE"])

    for region_id in (REGION_1, REGION_2, REGION_3):
        in_region = region_mask == region_id
        if region_id == REGION_3:
            bright_enough = luminance >= cfg["REGION_3_MIN_LUMINANCE"]
            effective_blackish = blackish | extra_blackish_region3
        else:
            bright_enough = np.ones((h, w), dtype=bool)
            effective_blackish = blackish
        good = in_region & opaque & ~effective_blackish & ~skeleton & bright_enough
        canvas[good] = arr[good]
        skipped_dark = int((in_region & effective_blackish).sum())
        skipped_dim = int((in_region & opaque & ~effective_blackish & ~skeleton & ~bright_enough).sum())
        print(f"  Region {region_id}: copied {good.sum()}, skipped {skipped_dark} blackish, {skipped_dim} too dim.")

    # Save intermediate (after step 3)
    os.makedirs(REMOVED_BLACK_DIR, exist_ok=True)
    inter_path = os.path.join(REMOVED_BLACK_DIR, basename)
    Image.fromarray(canvas, "RGBA").save(inter_path)
    print(f"Intermediate saved to {inter_path}")

    # Step 4: remove hue outliers
    print("Removing hue outliers...")
    remove_hue_outliers(canvas, region_mask, cfg)

    # Step 5: fill gaps
    print("Filling gaps...")
    gap_mask = (region_mask > 0) & (canvas[:, :, 3] == 0)  # transparent region pixels before fill
    for region_id in (REGION_1, REGION_2, REGION_3):
        weighted_random_fill(canvas, region_mask, region_id, skeleton, cfg)

    # Step 5.5: masked Gaussian blur on gap-filled pixels and
    # a few nearby copied good pixels (so avoid unnecessary blur but edge smooth).
    # Per-region so colors from adjacent regions or skeleton do not bleed in.

    # Standard deviation of the Gaussian kernel in pixels. Blur effectively
    # spreads ~2*sigma px from each pixel (where weight drops to ~14%).
    # Larger value: softer blend but may wash out the mottled texture.
    _GAUSS_SIGMA = 2

    # Copied pixels within this many pixels of a gap are also blurred,
    # so the transition is smoothed from both sides of the seam.
    # Should be >= _GAUSS_SIGMA so the blend zone covers the full kernel spread.
    _NEAR_GAP_PX = 4

    _near_gap_struct = np.ones((2 * _NEAR_GAP_PX + 1, 2 * _NEAR_GAP_PX + 1), dtype=bool)
    for region_id in (REGION_1, REGION_2, REGION_3):
        region_gap = gap_mask & (region_mask == region_id)
        if not np.any(region_gap):
            continue
        # copied pixels in this region that are within 4px of a gap pixel
        gap_dilated = binary_dilation(region_gap, structure=_near_gap_struct)
        near_gap_copied = (region_mask == region_id) & ~gap_mask & ~skeleton & gap_dilated
        write_mask = region_gap | near_gap_copied
        valid = (region_mask == region_id) & ~skeleton
        blurred_weights = gaussian_filter(valid.astype(np.float32), sigma=_GAUSS_SIGMA)
        for c in range(3):  # blur R, G, B; leave alpha unchanged
            values = np.where(valid, canvas[:, :, c].astype(np.float32), 0.0)
            blurred_values = gaussian_filter(values, sigma=_GAUSS_SIGMA)
            with np.errstate(invalid="ignore", divide="ignore"):
                normalized = np.where(blurred_weights > 0,
                                      blurred_values / blurred_weights,
                                      canvas[:, :, c].astype(np.float32))
            canvas[write_mask, c] = normalized[write_mask].clip(0, 255).astype(np.uint8)

    # Step 5.6: fill Bottom logo trapezoid with a vertical gradient.
    # Zone layout (by fraction of total height):
    #   0–35%  : solid black
    #   35–95% : linear ramp from black to region 1 mean color
    #   95–100%: solid region 1 mean color
    # Coordinates taken directly from the fallback.svg Bottom logo polyline.
    _LOGO_TOP_Y, _LOGO_BOT_Y = 1636, 1680
    _LOGO_TOP_XL, _LOGO_TOP_XR = 470, 710  # x bounds at top edge
    _LOGO_BOT_XL, _LOGO_BOT_XR = 426, 754  # x bounds at bottom edge
    _LOGO_RAMP_START = 0.35  # fraction where gradient begins (above is solid black)
    _LOGO_RAMP_SPAN = 0.60  # fraction length of the gradient ramp
    r1_pixels = canvas[(region_mask == REGION_1) & (canvas[:, :, 3] > 0)][:, :3]
    r1_color = r1_pixels.mean(axis=0) if len(r1_pixels) > 0 else np.array([180.0, 130.0, 60.0])
    logo_fill = np.zeros_like(canvas)
    total_h = _LOGO_BOT_Y - _LOGO_TOP_Y
    for y in range(_LOGO_TOP_Y, _LOGO_BOT_Y + 1):
        raw_t = (y - _LOGO_TOP_Y) / total_h
        t = float(np.clip((raw_t - _LOGO_RAMP_START) / _LOGO_RAMP_SPAN, 0.0, 1.0))
        x_l = round(_LOGO_TOP_XL + (_LOGO_BOT_XL - _LOGO_TOP_XL) * raw_t)
        x_r = round(_LOGO_TOP_XR + (_LOGO_BOT_XR - _LOGO_TOP_XR) * raw_t)
        rgb = (r1_color * t).clip(0, 255).astype(np.uint8)
        logo_fill[y, x_l:x_r, :3] = rgb
        logo_fill[y, x_l:x_r, 3] = 255
    logo_mask = (logo_fill[:, :, 3] > 0) & ~skeleton
    canvas[logo_mask] = logo_fill[logo_mask]

    # Step 6: add border strip, ATK DEF strip
    filled_img = Image.fromarray(canvas, "RGBA")
    add_border_strip(filled_img, output_path, cfg)


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
