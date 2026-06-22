import {test, expect} from "@playwright/test"

// fontCardName (web/font/YGOSmallCaps.ttf, copied from Master Duel) has a
// "numbersign" glyph drawn as a cent sign instead of a hash crosshatch.
// The fix renders "#" through fontCardNameHashFix instead, inside RenderCardName.
test("card name hashtag renders through fontCardNameHashFix, not the broken fontCardName glyph", async ({page}) => {
    // GIVEN the card editor is open
    const fileURL = "file://" + __dirname + "/../web/index.html"
    await page.goto(fileURL)
    await page.waitForLoadState("load")

    // WHEN a card with "#" in its name is loaded (card ID 4013,
    // "Winged Dragon, Guardian of the Fortress #1")
    await page.evaluate(() => {
        ;(window as any).loadDevTestCardByID("4013")
    })
    await page.waitForTimeout(500)

    // THEN the "#" character should be wrapped in an element using fontCardNameHashFix,
    // so it does not inherit fontCardName's broken numbersign glyph
    const hashFontFamily = await page.evaluate(() => {
        const root = document.getElementById("RenderCardName")
        if (!root) return "RenderCardName not found"
        const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
        let node: Node | null
        while ((node = walker.nextNode())) {
            if (node.textContent === "#") {
                return window.getComputedStyle(node.parentElement as Element).fontFamily
            }
        }
        return "# text node not found"
    })
    expect(hashFontFamily).toBe("fontCardNameHashFix")
})
