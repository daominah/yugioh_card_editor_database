# Fill Plan

## Input

- web/card_frame_rushduel/fallback.svg: the skeleton with black strokes defining the card frame edges.
- input_png/monster_normal.png: have color scheme we want, but was upscaled
  from a smaller resolution, so it has black pixels near edges that we want to avoid copying
  (the input_png is extracted from Konami Duel Links).

The 2 input images have the same dimensions, 1180x1720 pixels.

Per-file config is stored in the INPUT_CONFIGS map in fill_by_region.py, keyed by input basename.
Each entry can override any parameter from the default config.

## Goal

Coloring the skeleton fallback.svg as similarly to the monster_normal.png.

- 1st region: between Inner card boundary and Art box (yellowish)
- 2nd region: between Effect box outer and Effect box inner (light yellowish, same as 1st region)
- 3rd region: inside Effect box inner (light yellowish-white)

The key property to preserve: the texture is **non-uniform and mottled**. Colors
vary continuously across the region (lighter patches, darker amber patches, subtle
noise). The filled gaps must blend into this variation, not produce flat solid areas.

Output is stored in [web/card_frame_rushduel](../web/card_frame_rushduel).

## Steps

- **Step 1: Start**: blank transparent canvas with the same dimensions as the input images.

- **Step 2: Draw skeleton**: rasterize fallback.svg and paint its black stroke pixels
  as opaque black onto the output canvas. These become the visible frame lines in the output.

- **Step 3: Copy good pixels**: for each pixel in each region:

  - A pixel is "blackish" if its luminance is below a threshold. Pixels near SVG skeleton
    lines use a stricter threshold because upscaling bleeds dark stroke color near edges.
  - The art box top and bottom horizontal edges use a separate, larger near-edge radius
    because they accumulate more bleed than the sides.
  - Pixels inside REGION_3 but near its boundary also use the stricter threshold,
    catching dark bleed the skeleton distance alone misses.
  - REGION_3 pixels must also meet a minimum luminance to be copied.
  - Blackish pixels are skipped; their canvas position stays transparent.

  After this step, save an intermediate result to python/removed_black/.

- **Step 4: Remove hue outliers**: erase copied pixels whose hue deviates too far from
  the average hue of their region. Only pixels with sufficient saturation are checked;
  grey/black pixels have unstable hue and are left alone. Erased pixels become transparent
  and are filled in the next step. Can be disabled.

- **Step 5: Fill gaps**: some pixels in each region are still transparent because their
  source pixels were blackish or erased. Fill them using **weighted random patch sampling**:
  for each gap pixel, find the k nearest already-copied pixels in the same region, then
  randomly pick one weighted so closer pixels are more likely. This introduces controlled
  randomness so adjacent gap pixels sample slightly different source pixels, reproducing
  the natural color variation of the original texture.

  For REGION_2 specifically: if a gap pixel does not have enough filled REGION_2 pixels
  nearby (REGION_2 is a thin strip that loses most pixels to blackish filtering),
  fall back to sampling from REGION_1 instead.

- **Step 5.5: Blur region pixels**: apply a Gaussian blur to the RGB channels of all
  region pixels to smooth color transitions between originally-copied and gap-filled areas.
  Only region pixels are updated; the skeleton strokes are restored to solid black afterward.
  Can be disabled by setting sigma to 0.

- **Step 6: Add border strip**: composite the filled PNG with the border_strip.png overlay
  and save the result directly to web/card_frame_rushduel/, replacing the frame asset.
