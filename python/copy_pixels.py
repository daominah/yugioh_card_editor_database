"""
Steps 3 and 4 of fill_by_region:
  3. Copy pixels from source into the canvas, filtered by region and optionally by luminance.
  4. Remove hue outliers from copied pixels.
"""

import os

import numpy as np
from PIL import Image
from scipy.ndimage import distance_transform_edt

from svg_masks import REGION_1, REGION_2, REGION_3, _art_box_tb_mask, _inner_boundary_mask


def copy_pixels(canvas, arr, region_mask, skeleton, cfg, svg_path, removed_black_dir, basename):
    """
    Step 3: copy source pixels into canvas for each region.
    Step 4: remove hue outliers (skipped when COPY_ALL_IN_REGION is True).

    When COPY_ALL_IN_REGION is True, all opaque non-skeleton pixels are copied
    without any luminance filter. When False, dark/blackish pixels are excluded
    using luminance thresholds that tighten near SVG edges.
    """
    opaque = arr[:, :, 3] >= cfg["ALPHA_THRESHOLD"]

    if cfg["COPY_ALL_IN_REGION"]:
        print("COPY_ALL_IN_REGION=True: copying all opaque pixels, skipping Step 4.")
        _copy_all(canvas, arr, region_mask, skeleton, opaque)
    else:
        _copy_filtered(canvas, arr, region_mask, skeleton, opaque, cfg, svg_path)
        print("Removing hue outliers...")
        remove_hue_outliers(canvas, region_mask, cfg)

    _save_intermediate(canvas, basename, removed_black_dir)


def remove_hue_outliers(canvas, region_mask, cfg):
    """
    Erase copied pixels whose hue deviates more than OUTLIER_MAX_HUE_DEV degrees
    from the average hue of their region. Only pixels with saturation >= 20 are
    checked; low-saturation pixels have unstable hue and are left to the blackish
    filter instead. Erased pixels become transparent for the gap-fill step.
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


# ---------------------------------------------------------------------------
# Internal helpers
# ---------------------------------------------------------------------------

def _copy_all(canvas, arr, region_mask, skeleton, opaque):
    for region_id in (REGION_1, REGION_2, REGION_3):
        in_region = region_mask == region_id
        good = in_region & opaque & ~skeleton
        canvas[good] = arr[good]
        print(f"  Region {region_id}: copied {good.sum()} pixels.")


def _copy_filtered(canvas, arr, region_mask, skeleton, opaque, cfg, svg_path):
    h, w = arr.shape[:2]
    inner_boundary = _inner_boundary_mask(svg_path, w, h)
    art_box_tb = _art_box_tb_mask(svg_path, w, h)
    blackish = _is_blackish(arr, skeleton, inner_boundary, art_box_tb, cfg)

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


def _save_intermediate(canvas, basename, removed_black_dir):
    os.makedirs(removed_black_dir, exist_ok=True)
    inter_path = os.path.join(removed_black_dir, basename)
    Image.fromarray(canvas, "RGBA").save(inter_path)
    print(f"Intermediate saved to {inter_path}")
