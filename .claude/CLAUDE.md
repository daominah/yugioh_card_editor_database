# HTML Comments

- For elements with begin/end tags (`<div>`, `<svg>`, etc.):
  place describing comments inside, between the tags.
- For self-closing elements (`<img>`, `<input>`, etc.):
  place describing comments before the element.

# Pixel Values

- Use even numbers for all pixel values in CSS and SVG
  (points, positions, dimensions, stroke widths),
  so scaling down by 0.5 produces no half-pixel artifacts.

# Dev Server

- Run `go run cmd/main-yugioh-card-editor/main.go` to start the server at http://localhost:20808

# Testing

- Playwright tests live in `web_tests/` (including `package.json`,
  `package-lock.json`, `playwright.config.ts`, and `node_modules/`).
  Run tests with `cd web_tests && npx playwright test`.
