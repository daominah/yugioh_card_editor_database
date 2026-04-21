"""
SVG parsing and rasterization helpers for fill_by_region.

Provides boolean/integer masks derived from fallback.svg geometry:
  rasterize_skeleton  : black stroke pixels from the SVG
  create_region_mask  : per-pixel region label (1, 2, 3, or 0)
  _inner_boundary_mask: stroke of the "inner card boundary" rect only
  _art_box_tb_mask    : top and bottom horizontal segments of the art box
"""

import xml.etree.ElementTree as ET

import numpy as np
from PIL import Image, ImageDraw

# The 3 regions from fill_by_region_plan.md.
REGION_1 = 1  # between Inner card boundary and Art box (yellowish)
REGION_2 = 2  # between Effect box outer and Effect box inner (same tone as region 1)
REGION_3 = 3  # inside Effect box inner / text area (lighter, yellowish-white)


# ---------------------------------------------------------------------------
# Low-level geometry helpers
# ---------------------------------------------------------------------------

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


# ---------------------------------------------------------------------------
# Public mask builders
# ---------------------------------------------------------------------------

def rasterize_skeleton(png_path, width, height, cfg):
    """Return bool mask (H x W) where SVG black stroke/fill pixels are True.
    Loads from a pre-rasterized PNG for pixel-accurate stroke positions,
    avoiding the half-pixel inward shift that PIL's integer stroke math introduces."""
    img = Image.open(png_path).convert("RGBA")
    if img.size != (width, height):
        img = img.resize((width, height), Image.LANCZOS)
    arr = np.array(img)
    return (arr[:, :, 0] < 10) & (arr[:, :, 3] > 200)


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


def _level_badge_mask(svg_path, width, height):
    """Return bool mask (H x W) covering the full level badge circle disk.
    Used to erase the badge from the PNG-loaded skeleton when COPY_LEVEL_BADGE=False.
    Erases a solid disk (r + stroke_width) rather than just the stroke ring, so that
    all PNG badge pixels are covered regardless of PIL stroke-centering inaccuracy."""
    root, sx, sy = _svg_scale(svg_path, width, height)
    canvas = Image.new("L", (width, height), 0)
    draw = ImageDraw.Draw(canvas)
    for elem in root.iter():
        if elem.tag.split("}")[-1] != "circle":
            continue
        if _title_of(elem) != "level badge":
            continue
        cx = float(elem.get("cx"))
        cy = float(elem.get("cy"))
        r = float(elem.get("r"))
        sw = max(1, round(float(elem.get("stroke-width", "1")) * (sx + sy) / 2))
        outer = r + sw  # covers the full stroke band including the outer edge
        draw.ellipse([((cx - outer) * sx, (cy - outer) * sy),
                      ((cx + outer) * sx, (cy + outer) * sy)], fill=255)
    return np.array(canvas) > 0
