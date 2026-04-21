#!/usr/bin/env python3
"""Check pixel color at a given (x, y) coordinate in an output card frame PNG.

Usage:
    python script_checkoutput_point.py <x> <y> [image_path]

    x, y        : pixel coordinates from top-left
    image_path  : optional; defaults to web/card_frame_rushduel/monster_xyz.png
"""

import os
import sys

from PIL import Image


def check_point(x, y, image_path):
    img = Image.open(image_path).convert("RGBA")
    r, g, b, a = img.getpixel((x, y))
    print(f"({x}, {y})  R={r}  G={g}  B={b}  A={a}")


if __name__ == "__main__":
    points = []
    img_path = os.path.join(
        os.path.dirname(__file__),
        "..", "web", "card_frame_rushduel", "monster_xyz.png")
    # "..", "web", "card_frame_rushduel", "fallback.png")
    if len(sys.argv) >= 3:
        x = int(sys.argv[1])
        y = int(sys.argv[2])
        points.append((x, y))
        if len(sys.argv) >= 4:
            img_path = sys.argv[3]
    else:
        points = [
            (155, 83),
            (155, 84),
            (424, 83),
            (72, 1603),
            (72, 1605),
        ]

    for point in points:
        check_point(point[0], point[1], img_path)
