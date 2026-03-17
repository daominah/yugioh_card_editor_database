---
name: test-card-by-name
description: Open the card editor in Playwright, search for a card by name, click the first result, and screenshot the rendered card. Use when you need to visually verify how a card renders after code changes, or when the user asks to test a specific card.
---

# Test Card by Name

Visually verify a rendered Yu-Gi-Oh card using Playwright MCP.

## Arguments

- `$ARGUMENTS`: The card name to search for (e.g. "Maxx", "Blue-Eyes").
  If not provided, ask the user for a card name.

## Important: Playwright workarounds

The page uses `document.body.style.transform = scale(0.5)` which causes
Playwright's direct click and element screenshot actions to timeout.

- **Click/search**: use `browser_evaluate` with JavaScript, never direct
  Playwright click.
- **Screenshot**: element screenshots timeout (card is 1180x1720px at full
  scale). Instead, resize viewport to match the card at 0.5 scale (590x860),
  scroll to the card column, and take a viewport screenshot.

**Note**: this skill was written and tested on a Windows machine with display
scale 200% on a 3840x2160 screen. The scroll offset and viewport dimensions
may need adjustment on other display configurations.

## Steps

1. Navigate to the card editor:
   ```
   browser_navigate to file:///C:/Users/tungd/go/src/github.com/daominah/yugioh_card_editor/web/index.html
   ```

2. Use `browser_evaluate` to fill the search box and trigger search:
   ```js
   document.getElementById('SearchCardQuery').value = '$ARGUMENTS';
   SearchCardDatabase();
   ```

3. Use `browser_evaluate` to click the first search result:
   ```js
   document.querySelector('#SearchCardResult .searchRow').click();
   ```
   If no results found, tell the user and stop.

4. Screenshot the card only:
   a. `browser_resize` to width=590, height=860
   b. `browser_evaluate` to scroll to the card: `window.scrollTo(590, 0)`
   c. `browser_take_screenshot` (viewport, no element ref)
   d. `browser_resize` back to width=1920, height=1080

5. Present the screenshot to the user and note any rendering issues if visible.
