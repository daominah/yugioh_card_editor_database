"""
Step 5 of fill_by_region:
  5.   Fill remaining transparent region pixels using weighted random patch sampling.
  5.5  Blur the seam between gap-filled and copied pixels.
  5.6  Fill the Bottom logo trapezoid with a vertical gradient.
"""

import numpy as np
from scipy.ndimage import binary_dilation, distance_transform_edt, gaussian_filter, uniform_filter
from scipy.spatial import cKDTree

from svg_masks import REGION_1, REGION_2, REGION_3


def fill_gaps(canvas, region_mask, skeleton, cfg):
    """Steps 5, 5.5, and 5.6."""
    # gap_mask is computed before filling so the blur step knows which pixels were gaps.
    gap_mask = (region_mask > 0) & (canvas[:, :, 3] == 0)

    # Pixels within SAMPLE_EXCLUDE_EDGE_PX of the skeleton are excluded from the sample pool
    # so that stroke-bleed pixels near SVG lines do not propagate into transparent gaps.
    px = cfg["SAMPLE_EXCLUDE_EDGE_PX"]
    near_skeleton = distance_transform_edt(~skeleton) <= px if px > 0 else None

    print("Filling gaps...")
    for region_id in (REGION_1, REGION_2, REGION_3):
        _weighted_random_fill(canvas, region_mask, region_id, skeleton, near_skeleton, cfg)

    _blur_gap_seams(canvas, region_mask, skeleton, gap_mask)
    _fill_bottom_logo(canvas, region_mask, skeleton)


# ---------------------------------------------------------------------------
# Internal helpers
# ---------------------------------------------------------------------------

def _weighted_random_fill(canvas, region_mask, region_id, skeleton, near_skeleton, cfg):
    """Fill transparent pixels in region using weighted random sampling from filled pixels.
    Skeleton pixels and, when near_skeleton is provided, pixels within SAMPLE_EXCLUDE_EDGE_PX
    of the skeleton are excluded from the source pool.

    For REGION_2: each gap pixel checks how many filled REGION_2 pixels exist within a
    local 100x100 window. If >= 100, it samples from REGION_2 (spatially nearby, correct
    tone). If not, it falls back to use REGION_1 color."""

    alpha_threshold = cfg["ALPHA_THRESHOLD"]
    rng = np.random.default_rng(cfg["RANDOM_FILL_SEED"])
    in_region = region_mask == region_id
    opaque = canvas[:, :, 3] >= alpha_threshold

    # Exclude skeleton and near-skeleton pixels from the sample pool so that
    # stroke-bleed pixels near SVG lines cannot propagate into transparent gaps.
    if near_skeleton is not None:
        not_sample = skeleton | near_skeleton
    else:
        not_sample = skeleton

    filled_own = in_region & opaque & ~not_sample
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
    filled_r1_coords = np.argwhere((region_mask == REGION_1) & opaque & ~not_sample)

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


def _blur_gap_seams(canvas, region_mask, skeleton, gap_mask):
    """
    Step 5.5: masked Gaussian blur on gap-filled pixels and a few nearby copied
    pixels, so the transition is smoothed from both sides of the seam.
    Per-region so colors from adjacent regions or skeleton do not bleed in.
    """
    # Blur effectively spreads ~2*sigma px from each pixel (where weight drops to ~14%).
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
        # copied pixels in this region that are within _NEAR_GAP_PX of a gap pixel
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


def _fill_bottom_logo(canvas, region_mask, skeleton):
    """
    Step 5.6: fill the Bottom logo trapezoid with a vertical gradient.
    Zone layout (by fraction of total height):
      0-35%  : solid black
      35-95% : linear ramp from black to region 1 mean color
      95-100%: solid region 1 mean color
    Coordinates taken directly from the fallback.svg Bottom logo polyline.
    """
    _LOGO_TOP_Y, _LOGO_BOT_Y = 1636, 1680
    _LOGO_TOP_XL, _LOGO_TOP_XR = 470, 710   # x bounds at top edge
    _LOGO_BOT_XL, _LOGO_BOT_XR = 426, 754   # x bounds at bottom edge
    _LOGO_RAMP_START = 0.35   # fraction where gradient begins (above is solid black)
    _LOGO_RAMP_SPAN = 0.60    # fraction length of the gradient ramp

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
