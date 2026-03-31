# Rush Duel Layout Feature Design

Date: 2026-03-18

## Overview

Add a second card output layout based on Rush Duel card style.
The user can switch between Standard (current) and Rush Duel layouts via a radio button.
Both layouts share the same card data model (`GlobalCard`).

---

## UI Changes

### Layout Switcher

- Location: `colRight`, after the "Card art resolution" section
- Control: two radio buttons
  ```
  Card layout:
  ○ Standard   ○ Rush Duel
  ```
- Default: Standard
- Handler: `handleClickCardLayout(layout)` sets `GlobalCardLayout`, persists to `localStorage`, and re-renders
- On page load: restore `GlobalCardLayout` from `localStorage` (key: `StorageKeyCardLayout`) before first render
- `StorageKeyCardLayout = "StorageKeyCardLayout"` — matches the casing pattern of `StorageKeyScale` / `StorageKeyArtResolution`

### New Form Fields (Rush Duel mode only, hidden in Standard mode)

| Field             | Location in colLeft                     | HTML element                                                            |
|-------------------|-----------------------------------------|-------------------------------------------------------------------------|
| LEGEND checkbox   | Same line as the Art file input         | `<input type="checkbox" id="RushIsLegend">`                             |
| Maximum checkbox  | Abilities row, 7th checkbox after Union | `<input type="checkbox" name="MonsterAbilities" id="RushMaximum">`      |
| MAXIMUM ATK input | Below Abilities, shown when Max checked | `<input id="RushMaximumATK" size="4">` (gold ATK line on the Main card) |

All fields are hidden (`display: none`) by default and shown only when Rush Duel layout is active. `RushMaximumATK` is additionally hidden unless the `RushMaximum` checkbox is checked.

`RushMaximumATKWrap` is a wrapper `<span>` around the MAXIMUM ATK label+input. The `RushMaximum` checkbox `onclick` handler toggles `RushMaximumATKWrap.style.display` directly (not via `handleClickCardLayout`), so it shows/hides in real time as the user checks/unchecks Maximum.

Note: the current Abilities row has 6 checkboxes in 2 columns of 3. Adding Maximum as a 7th may require a layout adjustment — either a 3rd column (3+3+1) or a 4th column of 2 rows (e.g. 2+2+2+1). The exact CSS tweak is deferred to implementation.

---

## Architecture

### Global State

```js
// New global variable
let GlobalCardLayout = "standard"; // "standard" | "rushduel"
```

### Two Parallel Card Renderers in colMid

```html
<!-- existing, untouched -->
<div id="RenderCard" class="cRenderCard db"> ...</div>

<!-- new Rush Duel renderer, hidden by default -->
<div id="RenderCardRushduel" class="cRenderCardRushduel db" style="display:none"> ...</div>
```

`#RenderCardRushduel` has its own child elements (name, attribute, level badge, ATK/DEF badges, art, type line, LEGEND badge, footer).

### Layout Toggle Logic

```js
function handleClickCardLayout(layout) {
	GlobalCardLayout = layout;
	document.getElementById("RenderCard").style.display =
		layout === "standard" ? "" : "none";
	document.getElementById("RenderCardRushduel").style.display =
		layout === "rushduel" ? "" : "none";

	// Show/hide Rush Duel-only fields
	const rdOnly = document.querySelectorAll(".rushDuelOnly");
	rdOnly.forEach(el => el.style.display = layout === "rushduel" ? "" : "none");

	renderCard(); // re-render whichever is active
}
```

---

## Rush Duel Rendering

### Card Frame

- Rush Duel frame PNGs are stored in `web/card_frame_rushduel/`
- Files are named identically to their standard counterparts in `web/card_frame/` (e.g. `monster_normal.png`, `spell.png`)
- `renderRushDuelFrame()` loads the corresponding PNG from `card_frame_rushduel/` based on card type/subtype, the same way `renderCardFrame()` selects from `card_frame/`
- If a PNG is missing for a given subtype, fall back to `card_frame_rushduel/fallback.svg` (white background, black border lines)

### Child Elements of `#RenderCardRushduel`

| Element        | ID                | Notes                                                                                        |
|----------------|-------------------|----------------------------------------------------------------------------------------------|
| Card name      | `Rush_CardName`   | Larger, bolder font than standard                                                            |
| Attribute icon | `Rush_Attribute`  | Top-right, same icon set as standard                                                         |
| LEGEND badge   | `Rush_Legend`     | Gold badge top-left of art area; hidden if `RushIsLegend` unchecked                          |
| MAXIMUM ATK    | `Rush_MaximumATK` | Gold line above ATK/DEF; shown when `RushMaximum` ability set and `RushMaximumATK` non-empty |
| Art            | `Rush_Art`        | Taller art area, fills more of the card                                                      |
| Level badge    | `Rush_Level`      | Bottom-left of art: star icon + "LEVEL" label + number                                       |
| ATK badge      | `Rush_ATK`        | Inline colored badge below art (red background)                                              |
| DEF badge      | `Rush_DEF`        | Inline colored badge below art (blue background)                                             |
| Type line      | `Rush_TypeLine`   | Below art: `[Dragon/Normal]` same format as standard                                         |
| Effect text    | `Rush_Effect`     | In the text box below type line                                                              |
| Footer         | `Rush_Footer`     | "RUSH DUEL" logo bottom-center + set code bottom-right                                       |

### New JS Render Functions

| Function                    | Responsibility                                                                              |
|-----------------------------|---------------------------------------------------------------------------------------------|
| `renderRushDuelCard()`      | Orchestrator; calls all sub-functions below                                                 |
| `renderRushDuelFrame()`     | Loads PNG from `card_frame_rushduel/`; falls back to `fallback.svg` if subtype file missing |
| `renderRushDuelName()`      | Renders card name with Rush Duel font styling                                               |
| `renderRushDuelAttribute()` | Renders attribute icon (reuses standard icon assets)                                        |
| `renderRushDuelLevel()`     | Renders star badge with LEVEL label and number                                              |
| `renderRushDuelAtkDef()`    | Renders red ATK and blue DEF inline badges                                                  |
| `renderRushDuelLegend()`    | Shows/hides gold LEGEND badge based on `RushIsLegend`                                       |
| `renderRushDuelEffect()`    | Renders effect text with auto font-size                                                     |
| `renderRushDuelFooter()`    | Renders RUSH DUEL logo and set code                                                         |

`renderCard()` gains a delegation block at its start: if `GlobalCardLayout === "rushduel"`, it calls `renderRushDuelCard(card)` and returns early, leaving the standard rendering path untouched.

**Text fitting:** `fitTextOneLine()` already exists in `index.js` (~line 566) — reuse it for `renderRushDuelName()`. Do not create a duplicate.

**Text overflow measurement:** `scrollHeight` may not reliably detect overflow on absolutely positioned elements. A hidden `<div id="testRushTextWidth" class="cTestText">` inside `#RenderCardRushduel` is used as a measurement container for auto font-size logic in `renderRushDuelEffect()`, the same way the standard renderer uses a scratch div.

---

## Card Data Model Changes

### `GlobalCard` additions

```js
GlobalCard.RushIsLegend     // bool — LEGEND badge on Rush Duel card
GlobalCard.RushMaximumATK   // string — MAXIMUM ATK value shown in gold on the Main card (e.g. "3400")
```

All new Rush Duel-specific fields on `GlobalCard` use the `Rush` prefix.
`RushMaximum` (Maximum ability) is stored in `MonsterAbilities` as the string `"RushMaximum"` — it is the 7th ability and shows in the type line as "Maximum".
`RushIsLegend` is a top-level boolean (not an ability).
`RushMaximumATK` is a top-level string, only relevant when `MonsterAbilities` includes `"RushMaximum"`.

---

## Rendering Differences Reference

| Element         | Standard                         | Rush Duel                                                                            |
|-----------------|----------------------------------|--------------------------------------------------------------------------------------|
| Frame           | PNG per subtype in `card_frame/` | PNG per subtype in `card_frame_rushduel/`; CSS fallback if file missing              |
| Card name font  | Mid-size, standard weight        | Larger, bolder                                                                       |
| Level display   | Row of star icons across top     | Single star badge bottom-left of art with LEVEL + number                             |
| ATK/DEF display | Plain text line at bottom        | Colored inline badges below the art (red=ATK, blue=DEF)                              |
| LEGEND badge    | Not present                      | Gold badge top-left of art, conditional on `RushIsLegend`                            |
| MAXIMUM ATK     | Not present                      | Gold `MAXIMUM ATK XXXX` line above normal ATK/DEF, when `RushMaximum` ability is set |
| Type line       | `[Dragon/Normal]`                | Same format                                                                          |
| Footer          | Set code bottom-right            | "RUSH DUEL" logo bottom-center + set code bottom-right                               |

---
