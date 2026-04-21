import {test, expect} from "@playwright/test"

// The CSS sets .cRenderMonsterLevel width to 936px,
// so expected starWidth = Math.floor(936 / 12 - 2) = 76px.
const expectedStarWidth = "76px"

test("starWidth is correct when starting from Rush Duel then switching to Standard", async ({page}) => {
    // GIVEN: localStorage is set to rushduel so the page loads in Rush Duel mode
    const fileURL = "file://" + __dirname + "/../web/index.html"
    await page.addInitScript(() => {
        localStorage.setItem("StorageKeyCardLayout", "rushduel")
    })
    await page.goto(fileURL)
    await page.waitForLoadState("load")

    // WHEN: user switches to Standard layout
    await page.evaluate(() => {
        (document.getElementById("LayoutStandard") as HTMLInputElement).click()
    })
    await page.waitForTimeout(500)

    // THEN: the level star elements should have the correct width (76px)
    const starWidth = await page.evaluate(() => {
        const star = document.getElementById("StarWrap1")
        return star ? star.style.width : "element not found"
    })
    expect(starWidth).toBe(expectedStarWidth)
})

test("starWidth is correct when starting directly in Standard layout", async ({page}) => {
    // GIVEN: localStorage is set to standard
    const fileURL = "file://" + __dirname + "/../web/index.html"
    await page.addInitScript(() => {
        localStorage.setItem("StorageKeyCardLayout", "standard")
    })
    await page.goto(fileURL)
    await page.waitForLoadState("load")

    // THEN: the level star elements should have the correct width (76px)
    const starWidth = await page.evaluate(() => {
        const star = document.getElementById("StarWrap1")
        return star ? star.style.width : "element not found"
    })
    expect(starWidth).toBe(expectedStarWidth)
})
