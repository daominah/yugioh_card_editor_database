#!/usr/bin/env python3
"""
script_calc_luminance.py: Print luminance for one or more hex colors.
"""


def luminance(hex_color):
    h = hex_color.lstrip("#")
    r, g, b = int(h[0:2], 16), int(h[2:4], 16), int(h[4:6], 16)
    lum = 0.299 * r + 0.587 * g + 0.114 * b
    return r, g, b, lum


if __name__ == "__main__":
    colorHex = "b4bcd1"
    r, g, b, lum = luminance(colorHex)
    print(f"#{colorHex.lstrip('#').upper()}  rgb({r}, {g}, {b})  luminance={lum:.1f}")
