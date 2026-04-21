"""
Step 6 of fill_by_region: composite the filled canvas with the border strip
and ATK/DEF strip.
"""

import io

from PIL import Image


def add_border_strip(base_img, cfg, atk_def_strip_svg, border_strip_png):
    """Composite the filled image with the border strip and ATK/DEF dark
    strip overlays. Returns the composited PIL Image."""
    w, h = base_img.size

    if cfg["ADD_ATK_DEF_STRIP"]:
        # Rasterize atk_def_strip.svg directly (single source of truth
        # for shape, color, and opacity).
        import cairosvg
        png_data = cairosvg.svg2png(url=atk_def_strip_svg, output_width=w, output_height=h)
        strip_layer = Image.open(io.BytesIO(png_data)).convert("RGBA")
        base_img = Image.alpha_composite(base_img, strip_layer)

    border = Image.open(border_strip_png).convert("RGBA").resize(base_img.size, Image.LANCZOS)
    return Image.alpha_composite(base_img, border)
