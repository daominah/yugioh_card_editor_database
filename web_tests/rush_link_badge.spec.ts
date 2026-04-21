import {test, expect} from "@playwright/test"

test("Rush Duel Link badge shows arrows", async ({page}) => {
	// GIVEN the card editor is open at localhost
	await page.goto("http://localhost:20808/")
	await page.waitForTimeout(1000)

	// WHEN we switch to Rush Duel layout
	await page.evaluate(() => {
		;(document.getElementById("LayoutRushduel") as HTMLInputElement).click()
	})
	await page.waitForTimeout(500)

	// AND load a Link4 monster (Apollousa) via dev button
	await page.evaluate(() => {
		;(window as any).loadDevTestCardByID("14496")
	})
	await page.waitForTimeout(1000)

	// THEN the link badge SVG should be visible
	let badgeVisible = await page.evaluate(() => {
		let badge = document.getElementById("Rush_LinkBadge")
		return badge !== null && badge.style.display !== "none"
	})
	expect(badgeVisible).toBe(true)

	// AND the correct 4 arrows should be visible (Up, Down, DownLeft, DownRight)
	let visibleArrows = await page.evaluate(() => {
		let arrows = [
			"RushArrowUp",
			"RushArrowDown",
			"RushArrowLeft",
			"RushArrowRight",
			"RushArrowUpLeft",
			"RushArrowUpRight",
			"RushArrowDownLeft",
			"RushArrowDownRight",
		]
		let visible: string[] = []
		for (let id of arrows) {
			let el = document.getElementById(id)
			if (el && el.style.visibility === "visible") {
				visible.push(id)
			}
		}
		return visible
	})
	expect(visibleArrows).toEqual([
		"RushArrowUp",
		"RushArrowDown",
		"RushArrowDownLeft",
		"RushArrowDownRight",
	])
})
