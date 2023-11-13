// byId is a shorthand for document.getElementById,
function byId(e) { return document.getElementById(e) }


let CardType = {
	Monster: "Monster",
	Spell: "Spell",
	Trap: "Trap",
}

let CardSubtype = {
	MonsterNormal: "MonsterNormal",
	MonsterEffect: "MonsterEffect",
	MonsterRitual: "MonsterRitual",
	MonsterFusion: "MonsterFusion",
	MonsterSynchro: "MonsterSynchro",
	MonsterXyz: "MonsterXyz",
	MonsterLink: "MonsterLink",

	SpellNormal: "SpellNormal",
	SpellQuickPlay: "SpellQuickPlay",
	SpellRitual: "SpellRitual",
	SpellContinuous: "SpellContinuous",
	SpellField: "SpellField",
	SpellEquip: "SpellEquip",

	TrapNormal: "TrapNormal",
	TrapCounter: "TrapCounter",
	TrapContinuous: "TrapContinuous",
}

let MonsterAttribute = {
	DARK: "DARK",
	EARTH: "EARTH",
	FIRE: "FIRE",
	LIGHT: "LIGHT",
	WATER: "WATER",
	WIND: "WIND",
	DIVINE: "DIVINE",
}

let Ability = {
	Flip: "Flip",
	Gemini: "Gemini",
	Spirit: "Spirit",
	Toon: "Toon",
	Tuner: "Tuner",
	Union: "Union",
}

let LinkArrow = {
	UpLeft: "UpLeft",
	Up: "Up",
	UpRight: "UpRight",
	Left: "Left",
	Right: "Right",
	DownLeft: "DownLeft",
	Down: "Down",
	DownRight: "DownRight",
}

// MonsterType has 25 types (excluded "Illusion"),
// https://yugipedia.com/wiki/Type
let MonsterType = {
	Aqua: "Aqua",
	Beast: "Beast",
	BeastWarrior: "Beast-Warrior",
	CreatorGod: "Creator God",
	Cyberse: "Cyberse",
	Dinosaur: "Dinosaur",
	DivineBeast: "Divine-Beast",
	Dragon: "Dragon",
	Fairy: "Fairy",
	Fiend: "Fiend",
	Fish: "Fish",
	Illusion: "Illusion",
	Insect: "Insect",
	Machine: "Machine",
	Plant: "Plant",
	Psychic: "Psychic",
	Pyro: "Pyro",
	Reptile: "Reptile",
	Rock: "Rock",
	SeaSerpent: "Sea Serpent",
	Spellcaster: "Spellcaster",
	Thunder: "Thunder",
	Warrior: "Warrior",
	WingedBeast: "Winged Beast",
	Wyrm: "Wyrm",
	Zombie: "Zombie",
}


// _____________________________________________________________________________
// _____________________________________________________________________________

const testCNameL03 = "Umi"
const testCNameL08 = 'Maxx "C"'
const testCNameL19 = "PSY-Framegear Gamma"
const testCNameL22 = "Blue-Eyes White Dragon"
const testCNameL36 = "Divine Arsenal AA-ZEUS - Sky Thunder"
const testCNameL47 = "Number 38: Hope Harbinger Dragon Titanic Galaxy"
const testCNameL52 = "Black Luster Soldier - Envoy of the Evening Twilight"

const testCEffectL012ST = `Draw 1 card.`
const testCEffectL046ST = `① Destroy all monsters your opponent controls.`
const testCEffectL069ST = `① Add 1 Level 4 or lower Warrior monster from your Deck to your hand.`
const testCEffectL123ME = `1 Tuner + 1+ non-Tuner monsters
① Your opponent cannot target this card with card effects, except during your Main Phase 2.`
const testCEffectL330ST = `If you control no cards, you can activate this card from your hand.
① Target 1 face-up monster your opponent controls; negate its effects (until the end of this turn), then, if this card was Set before activation and is on the field at resolution, for the rest of this turn all other Spell/Trap effects in this column are negated.`
const testCEffectL264ME = `1 Level 1 monster, except a Token
You can only use the effect ① of "Relinquished Anima" once per turn.
① You can target 1 face-up monster this card points to; equip that face-up monster to this card (max. 1).
② This card gains ATK equal to that equipped monster's.`
const testCEffectL276ME = `2 Level 4 monsters
① While this card has a material attached that was originally WATER, all WATER monsters you control gain 500 ATK.
② Once per turn (Quick Effect): You can detach 1 material from this card; your opponent cannot activate any card effects in their GY this turn.`
const testCEffectL508ME = `1 Tuner + 1 or more non-Tuner Synchro Monsters
① Once per turn, when another monster's effect is activated (Quick Effect): You can negate the activation, and if you do, destroy it, and if you do that, this card gains ATK equal to the destroyed monster's original ATK until the end of this turn.
② If this card battles an opponent's Level 5 or higher monster, during damage calculation: This card gains ATK equal to the current ATK of the opponent's monster it is battling during that damage calculation only.`
const testCEffectL514ME = `2 Level 12 monsters
Once per turn, if an Xyz Monster battled this turn, you can also Xyz Summon "Divine Arsenal AA-ZEUS - Sky Thunder" by using 1 Xyz Monster you control as material. (Transfer its materials to this card.)
① (Quick Effect): You can detach 2 materials from this card; send all other cards from the field to the GY.
② Once per turn, if another card(s) you control is destroyed by battle or an opponent's card effect: You can attach 1 card from your hand, Deck, or Extra Deck to this card as material.`
const testCEffectL571MM = `You can only use each of the following effects ① and ③ of "Blue-Eyes Jet Dragon" once per turn, and can only activate them while "Blue-Eyes White Dragon" is on your field or in your GY.
① If a card(s) on the field is destroyed by battle or card effect: You can Special Summon this card from the GY (if it was there when the card was destroyed) or hand (even if not).
② Other cards you control cannot be destroyed by your opponent's card effects.
③ At the start of the Damage Step, if this card battles: You can target 1 card your opponent controls; return it to the hand.`
const testCEffectL594ME = `2 Level 6 monsters
You can also Xyz Summon this card by sending 1 "Burning Abyss" monster from your hand to the GY, then using 1 "Dante" monster you control as material. (Transfer its materials to this card.) If Summoned this way, the following effect ① cannot be activated this turn.
① Once per turn (Quick Effect): You can detach 1 material from this card; send 1 card from your Deck to the GY.
② If this card in your possession is destroyed by your opponent's card and sent to your GY: You can Special Summon 1 "Burning Abyss" monster from your Extra Deck, ignoring its Summoning conditions.`
const testCEffectL630MM = `You can only use 1 "Lord of the Heavenly Prison" effect per turn, and only once that turn.
① During your Main Phase: You can activate this effect; this card in your hand becomes revealed until the end of your opponent's turn. While this card is revealed by this effect, Set cards on the field cannot be destroyed by card effects.
② If a Set Spell/Trap Card is activated (except during the Damage Step): You can Special Summon this card from your hand, then, if you activated this effect while this card was revealed, you can reveal and Set 1 Spell/Trap directly from your Deck, but banish it during the End Phase of the next turn.`
const testCEffectL579PM = `① Once per turn, when a Spell/Trap Card or effect is activated (Quick Effect): You can return 1 card you control with a Spell Counter to the hand, and if you do, negate the activation, and if you do that, destroy it. Then, you can place the same number of Spell Counters on this card that the returned card had.
② While this card has a Spell Counter, your opponent cannot target it with card effects, also it cannot be destroyed by your opponent's card effects.
③ When this card with a Spell Counter is destroyed by battle: You can add 1 Normal Spell from your Deck to your hand.`
const testCEffectL400PP = `You can only use the Pendulum Effect of "Endymion, the Mighty Master of Magic" once per turn.
① You can remove 6 Spell Counters from your field; Special Summon this card from the Pendulum Zone, then count the number of cards you control that can have a Spell Counter, destroy up to that many cards on the field, and if you do, place Spell Counters on this card equal to the number of cards destroyed.`

// _____________________________________________________________________________
// _____________________________________________________________________________

let DefaultCard = {
	CardName: "",
	CardType: CardType.Monster,
	CardSubtype: CardSubtype.MonsterNormal,
	CardEffect: "",
	CardArt: "",

	MonsterAttribute: MonsterAttribute.DARK,
	MonsterType: MonsterType.Warrior,
	MonsterLevelRankLink: 0,
	MonsterATK: 0,
	MonsterDEF: 0,
	MonsterAbilities: [],  // Ability.Tuner, Ability.Flip, ...
	MonsterLinkArrows: [],  // LinkArrow.Up, LinkArrow.UpRight, ...

	IsPendulum: false,
	PendulumScale: 0,
	PendulumEffect: "",

	MiscKonamiSet: "",
	MiscKonamiCardID: "",
	MiscYear: (new Date()).getFullYear(),
	MiscCreator: "daominah",
}

let DefaultCardZeus = {
// DefaultCard = {
	CardName: testCNameL36,
	CardType: CardType.Monster,
	CardSubtype: CardSubtype.MonsterXyz,
	CardEffect: testCEffectL514ME,
	CardArt: "card/divine_zeus_15524.jpg",

	MonsterAttribute: MonsterAttribute.LIGHT,
	MonsterType: MonsterType.Machine,
	MonsterLevelRankLink: 12,
	MonsterATK: 3000,
	MonsterDEF: 3000,
	MonsterAbilities: [],  // Ability.Tuner, Ability.Flip, ...
	MonsterLinkArrows: [],  // LinkArrow.Up, LinkArrow.UpRight, ...

	IsPendulum: false,
	PendulumScale: 0,
	PendulumEffect: "",

	MiscKonamiSet: "PHRA-EN045",
	MiscKonamiCardID: "15524",
	MiscYear: 2020,
	MiscCreator: "daominah",
}

let DefaultCardJet = {
// DefaultCard = {
	CardName: "Blue-Eyes Jet Dragon",
	CardType: CardType.Monster,
	CardSubtype: CardSubtype.MonsterEffect,
	CardEffect: testCEffectL571MM,
	CardArt: "card/blue_eyes_jet_16809.jpg",

	MonsterAttribute: MonsterAttribute.LIGHT,
	MonsterType: MonsterType.Dragon,
	MonsterLevelRankLink: 8,
	MonsterATK: 3000,
	MonsterDEF: 0,
	MonsterAbilities: [],  // Ability.Tuner, Ability.Flip, ...
	MonsterLinkArrows: [],  // LinkArrow.Up, LinkArrow.UpRight, ...

	IsPendulum: false,
	PendulumScale: 0,
	PendulumEffect: "",

	MiscKonamiSet: "BACH-EN004",
	MiscKonamiCardID: "16809",
	MiscYear: 2022,
	MiscCreator: "daominah",
}

let DefaultCardRelinq = {
// DefaultCard = {
	CardName: "Relinquished Anima",
	CardType: CardType.Monster,
	CardSubtype: CardSubtype.MonsterLink,
	CardEffect: testCEffectL264ME,
	CardArt: "card/relinquished_anima_13841.jpg",

	MonsterAttribute: MonsterAttribute.DARK,
	MonsterType: MonsterType.Spellcaster,
	MonsterLevelRankLink: 2,
	MonsterATK: 3000,
	MonsterDEF: 2500,
	MonsterAbilities: [],  // Ability.Tuner, Ability.Flip, ...
	MonsterLinkArrows: [LinkArrow.Up],  // LinkArrow.Up, LinkArrow.UpRight, ...

	IsPendulum: false,
	PendulumScale: 0,
	PendulumEffect: "",

	MiscKonamiSet: "DUOV-EN053",
	MiscKonamiCardID: "13841",
	MiscYear: 2020,
	MiscCreator: "daominah",
}


// _____________________________________________________________________________
// _____________________________________________________________________________


let MapImg = {
	Spell: "icon/attr_SPELL.png",
	Trap: "icon/attr_TRAP.png",

	DARK: "icon/attr_DARK.png",
	DIVINE: "icon/attr_DIVINE.png",
	EARTH: "icon/attr_EARTH.png",
	FIRE: "icon/attr_FIRE.png",
	LIGHT: "icon/attr_LIGHT.png",
	WATER: "icon/attr_WATER.png",
	WIND: "icon/attr_WIND.png",

	TrapCounter: "icon/GUI_T_Icon1_Icon01.png",
	SpellField: "icon/GUI_T_Icon1_Icon02.png",
	SpellEquip: "icon/GUI_T_Icon1_Icon03.png",
	TrapContinuous: "icon/GUI_T_Icon1_Icon04.png",
	SpellContinuous: "icon/GUI_T_Icon1_Icon04.png",
	SpellQuickPlay: "icon/GUI_T_Icon1_Icon05.png",
	SpellRitual: "icon/GUI_T_Icon1_Icon06.png",

	Level: "icon/GUI_T_Icon1_Other_Level.png",
	Rank: "icon/GUI_T_Icon1_Other_Rank.png",
}

let MapCardSubtypeText = {
	SpellNormal: "Normal Spell",
	SpellQuickPlay: "Quick-Play Spell",
	SpellRitual: "Ritual Spell",
	SpellContinuous: "Continuous Spell",
	SpellField: "Field Spell",
	SpellEquip: "Equip Spell",
	TrapNormal: "Normal Trap",
	TrapCounter: "Counter Trap",
	TrapContinuous: "Continuous Trap",
}

// _____________________________________________________________________________
// _____________________________________________________________________________

// GlobalCard's value will be updated by colLeft inputs,
// colMid will use the GlobalCard to render the card image,
// this var scope is global for easier debug, can be removed.
let GlobalCard = NewCard();

let LastUpdateCardState = new Date(0)

// updateCardState reads "colLeft" then draws to "colMid",
// this function will be called when anything on "colLeft" changed, multiple
// calls to this function do not update HTML more than once per 100ms
function updateCardState() {
	let now = new Date()
	let sinceLast = now - LastUpdateCardState // milliseconds
	if (sinceLast < 100) {
		// console.log(`${now.toISOString()} SKIP updateCardState, sinceLast: ${sinceLast} ms`)
		return
	}
	LastUpdateCardState = now
	// console.log(`${now.toISOString()} updateCardState, sinceLast: ${sinceLast} ms`)
	GlobalCard = readCardFromHTML();
	renderCard(GlobalCard)
}

// handleClickCardType shows corresponding CardSubtype elements of the cardType
function handleClickCardType(cardType) {
	// console.log("handleClickCardType:", cardType)
	let em = byId("CardSubtypeMonster")
	let es = byId("CardSubtypeSpell")
	let et = byId("CardSubtypeTrap")
	let show = em
	switch (cardType) {
		case CardType.Monster:
			em.style.display = ""
			es.style.display = "none"
			et.style.display = "none"
			break
		case CardType.Spell:
			em.style.display = "none"
			es.style.display = ""
			et.style.display = "none"
			show = es
			break
		default: // case CardType.Trap:
			em.style.display = "none"
			es.style.display = "none"
			et.style.display = ""
			show = et
	}
	show.getElementsByTagName("input")[0].checked = true
	if (cardType === CardType.Monster) {
		byId("MonsterDetail").classList.remove("disabledElement")
	} else {
		byId("MonsterDetail").classList.add("disabledElement")
		byId("IsPendulum").checked = false
	}
	updateCardState()
}

function handleInput(event) {
	// console.log(`event input tag: ${event.target.tagName}, id: ${event.target.id}`)
	updateCardState()
}

function handleClick() {
	// console.log(`event click tag: ${this.id}`)
	updateCardState()
}


function loadMonsterTypeElements() {
	let select = byId("MonsterType")
	let monsterTypes = Object.keys(MonsterType)
	for (let i = 0; i < monsterTypes.length; i++) {
		let option = document.createElement("option")
		option.id = MonsterType[monsterTypes[i]]
		option.value = MonsterType[monsterTypes[i]]
		option.innerHTML = MonsterType[monsterTypes[i]]
		select.appendChild(option)
	}
}


function hideMonsterDetailElements() {
	byId("RenderMonsterAbilities").style.display = "none"
	byId("RenderMonsterAbilitiesSmall").style.display = "none"
	byId("RenderMonsterEffect").style.display = "none"
	byId("RenderMonsterEffectSmall").style.display = "none"
	byId("RenderMonsterSplitLine").style.display = "none"
	byId("RenderMonsterATKLabel").style.display = "none"
	byId("RenderMonsterATK").style.display = "none"
	byId("RenderMonsterDEFLabel").style.display = "none"
	byId("RenderMonsterDEF").style.display = "none"
	byId("RenderMonsterLinkLabel").style.display = "none"
	byId("RenderMonsterLinkRating").style.display = "none"
}

// calcTextWidth calculates the text width on one line nowrap; by rendering
// the text on a SHARED hidden element then measure clientWidth; this is kind of
// a hack, I think concurrent calls to this function may cause wrong results but
// things seem right (is being used in fitTextOneLine to render CardName,
// MonsterAbilities, MonsterAtkDefLink)
function calcTextWidth(text, styleFont) {
	let test = byId("testTextWidth")
	test.innerHTML = text
	test.style.whiteSpace = "nowrap"
	test.style.font = styleFont
	let width = test.clientWidth
	test.innerHTML = ""
	// console.log(`calcTextWidth "${text}" ${styleFont} result: ${width}`)
	return width;
}

// renderTextFitOneLine clears the input element then fit the text into it,
// the text width will be scaled automatically if overflowed,
// magic value scaleFont=1.5 and scaleH=1.15 helps to fit card name
function fitTextOneLine(text, element, scaleFont = 1.0,
						scaleH = 1.15, scaleW = 1.0) {
	if (!element) {
		console.log(`error fitTextOneLine element: ${element}, should be unreachable`)
		return
	}
	element.innerHTML = ""
	element.style.lineHeight = "1"
	element.style.fontSize = (element.offsetHeight * scaleFont).toString() + "px"
	let textW = calcTextWidth(text, window.getComputedStyle(element).font)
	if (textW > element.clientWidth) {
		scaleW = element.clientWidth / textW
	}
	// console.log(`fitTextOneLine ${element.id} scaleW: ${scaleW}`)
	let child = document.createElement("div")
	child.textContent = text
	child.style.transform = `scale(${scaleW}, ${scaleH})`
	child.style.transformOrigin = "bottom left"
	let magicBaseline = document.createElement("div")
	magicBaseline.style.height = element.offsetHeight.toString() + "px"
	magicBaseline.style.display = "inline-block"
	child.appendChild(magicBaseline)
	element.appendChild(child)
}

// chooseFont renders the text on a SHARED hidden element then measures
// the clientHeight, if overflowed, repeat with a smaller fontSize,
// this function return an integer (need to add "px" to set fontSize)
function chooseFontSize(textHTML, width, height, fontFamily, fontWeight) {
	let test = byId("testTextHeight")
	test.style.width = `${width}px`
	test.style.fontFamily = fontFamily
	test.style.fontWeight = fontWeight
	let chosenSize = 42
	let log = "";
	for (let i = 0; i < 20; i++) {
		test.innerHTML = textHTML
		test.style.fontSize = `${chosenSize}px`
		log = `chooseFontSize ${chosenSize}px, testWH: ${test.clientWidth}x${test.clientHeight}, targetWH: ${width}x${height}`
		// console.log(log)
		if (test.clientHeight <= height) {
			break
		}
		chosenSize -= 1
	}
	console.log(log)
	test.innerHTML = ""
	return chosenSize
}


// NewCard returns a card obj with default fields value
function NewCard() {
	let jsoned = JSON.stringify(DefaultCard)
	return JSON.parse(jsoned)
}

function CloneCard(card) {
	return JSON.parse(JSON.stringify(card))
}


// readCardFromHTML returns a card object with all fields value read from
// current page HTML "colLeft", if a field is undefined or null, this function
// will set the field value equals to DefaultCard
function readCardFromHTML() {
	let c = NewCard()
	c.CardName = byId("CardName").value
	if (c.CardName) {
		c.CardName = c.CardName.trim()
	} else {
		c.CardName = DefaultCard.CardName
	}

	let radiosCardType = document.querySelector('input[name="CardType"]:checked')
	if (radiosCardType) {
		c.CardType = radiosCardType.id
	}
	let radiosCardSubtype = document.querySelector('input[name="CardSubtype"]:checked')
	if (radiosCardSubtype) {
		c.CardSubtype = radiosCardSubtype.id
	}

	c.CardEffect = byId("CardEffect").value
	if (c.CardEffect) {
		c.CardEffect = c.CardEffect.trim()
	} else {
		c.CardEffect = DefaultCard.CardEffect
	}

	let radiosMonsterAttribute = document.querySelector(
		'input[name="MonsterAttribute"]:checked')
	if (radiosMonsterAttribute) {
		c.MonsterAttribute = radiosMonsterAttribute.id
	}

	c.MonsterType = byId("MonsterType").value
	if (!c.MonsterType) {
		c.MonsterType = DefaultCard.MonsterType
	}

	c.MonsterLevelRankLink = byId("MonsterLevelRankLink").value
	if (c.MonsterLevelRankLink) {
		c.MonsterLevelRankLink = Math.floor(c.MonsterLevelRankLink)
	} else {
		c.MonsterLevelRankLink = DefaultCard.MonsterLevelRankLink
	}
	c.MonsterATK = byId("MonsterATK").value
	if (c.MonsterATK) {
		c.MonsterATK = Math.floor(c.MonsterATK)
	} else {
		c.MonsterATK = DefaultCard.MonsterATK
	}
	c.MonsterDEF = byId("MonsterDEF").value
	if (c.MonsterDEF) {
		c.MonsterDEF = Math.floor(c.MonsterDEF)
	} else {
		c.MonsterDEF = DefaultCard.MonsterDEF
	}

	c.MonsterAbilities = []
	let abilitiesInputs = document.getElementsByName("MonsterAbilities")
	for (let i = 0; i < abilitiesInputs.length; i++) {
		let e = abilitiesInputs[i];
		if (e.checked) {
			c.MonsterAbilities.push(e.id)
		}
	}
	c.MonsterAbilities.sort()

	c.MonsterLinkArrows = []
	let linkArrowInputs = document.getElementsByName("MonsterLinkArrows")
	for (let i = 0; i < linkArrowInputs.length; i++) {
		let e = linkArrowInputs[i];
		if (e.checked) {
			c.MonsterLinkArrows.push(e.id)
		}
	}

	c.IsPendulum = byId("IsPendulum").checked
	if (c.IsPendulum) {
		c.PendulumScale = byId("PendulumScale").value
		if (c.PendulumScale) {
			c.PendulumScale = Math.floor(c.PendulumScale)
		} else {
			c.PendulumScale = DefaultCard.PendulumScale
		}
		c.PendulumEffect = byId("PendulumEffect").value
		if (c.PendulumEffect) {
			c.PendulumEffect = c.PendulumEffect.trim()
		} else {
			c.PendulumEffect = DefaultCard.PendulumEffect
		}
	}
	if (c.IsPendulum) {
		c.CardArt = byId("ImgRenderCardArtPendulum").src
	} else {
		c.CardArt = byId("ImgRenderCardArt").src
	}

	c.MiscKonamiSet = byId("SetNumber").value
	c.MiscKonamiCardID = byId("CardID").value
	c.MiscYear = byId("Year").value
	c.MiscCreator = byId("Creator").value

	return c
}

// loadCardToHTML uses the input card object to fill HTML "colLeft" elements
function loadCardToHTML(c) {
	byId("CardName").value = c.CardName
	byId(c.CardType).checked = true
	let em = byId("CardSubtypeMonster")
	let es = byId("CardSubtypeSpell")
	let et = byId("CardSubtypeTrap")
	switch (c.CardType) {
		case CardType.Monster:
			em.style.display = ""
			es.style.display = "none"
			et.style.display = "none"
			break
		case CardType.Spell:
			em.style.display = "none"
			es.style.display = ""
			et.style.display = "none"
			break
		default: // case CardType.Trap:
			em.style.display = "none"
			es.style.display = "none"
			et.style.display = ""
	}
	byId(c.CardSubtype).checked = true
	if (c.CardArt) {
		byId("ImgRenderCardArtPendulum").src = c.CardArt
		byId("ImgRenderCardArt").src = c.CardArt
	}
	byId("CardEffect").value = c.CardEffect
	if (byId(c.MonsterAttribute)) {
		byId(c.MonsterAttribute).checked = true
	}
	byId("MonsterType").value = c.MonsterType
	byId("MonsterLevelRankLink").value = c.MonsterLevelRankLink
	byId("MonsterATK").value = c.MonsterATK
	byId("MonsterDEF").value = c.MonsterDEF

	for (let k in Ability) {
		let checkbox = byId(k)
		if (checkbox) {
			checkbox.checked = false
		}
	}
	if (c.MonsterAbilities) {
		for (let i = 0; i < c.MonsterAbilities.length; i++) {
			let checkbox = byId(c.MonsterAbilities[i])
			if (checkbox) {
				checkbox.checked = true
			}
		}
	}

	for (let k in LinkArrow) {
		let checkbox = byId(k)
		if (checkbox) {
			checkbox.checked = false
		}
	}
	if (c.MonsterLinkArrows) {
		for (let i = 0; i < c.MonsterLinkArrows.length; i++) {
			let checkbox = byId(c.MonsterLinkArrows[i])
			if (checkbox) {
				checkbox.checked = true
			}
		}
	}

	byId("IsPendulum").checked = !!c.IsPendulum;
	if (c.PendulumScale) {
		byId("PendulumScale").value = c.PendulumScale
	}
	if (c.PendulumEffect) {
		byId("PendulumEffect").value = c.PendulumEffect
	}

	byId("SetNumber").value = c.MiscKonamiSet
	byId("CardID").value = c.MiscKonamiCardID
	byId("Year").value = c.MiscYear
	if (c.MiscCreator) {  // keep "Creator"
		byId("Creator").value = c.MiscCreator
	}

	return c
}


// renderCard draw the card image by updating HTML "colMid"
function renderCard(card) {
	renderCardFrame(card)
	renderCardName(card)
	renderCardAttribute(card)
	renderCardTypeLevelRank(card)
	renderLinkArrow(card)
	renderMisc(card)
	renderPendulum(card)

	let [chosenEffectElement, autoFontSize] = renderCardEffect(card)
	byId("AutoFont").value = autoFontSize
	byId("ChosenEffectElementID").value = chosenEffectElement.id
}

function renderCardFrame(card) {
	let s = byId("RenderCard").style
	if (card.CardType === CardType.Spell) {
		s.backgroundImage = "url(card_frame/spell.png)"
		//
		// s.backgroundImage = "url(example/eg_spell_normal.jpg)"
		// s.backgroundImage = "url(example/eg_spell_field.jpg)"
		// s.backgroundImage = "url(example/eg_spell_ritual.jpg)"
	} else if (card.CardType === CardType.Trap) {
		s.backgroundImage = "url(card_frame/trap.png)"
		//
		// s.backgroundImage = "url(example/eg_trap_normal.jpg)"
		// s.backgroundImage = "url(example/eg_trap_counter.jpg)"
		// s.backgroundImage = "url(example/eg_trap_continuous.jpg)"
	} else {  // CardType.Monster
		if (!card.IsPendulum) {
			switch (card.CardSubtype) {
				case CardSubtype.MonsterNormal:
					s.backgroundImage = "url(card_frame/monster_normal.png)"
					break
				case CardSubtype.MonsterEffect:
					s.backgroundImage = "url(card_frame/monster_effect.png)"
					break
				case CardSubtype.MonsterRitual:
					s.backgroundImage = "url(card_frame/monster_ritual.png)"
					break
				case CardSubtype.MonsterFusion:
					s.backgroundImage = "url(card_frame/monster_fusion.png)"
					break
				case CardSubtype.MonsterSynchro:
					s.backgroundImage = "url(card_frame/monster_synchro.png)"
					break
				case CardSubtype.MonsterXyz:
					s.backgroundImage = "url(card_frame/monster_xyz.png)"
					break
				case CardSubtype.MonsterLink:
					s.backgroundImage = "url(card_frame/monster_link.png)"
					break
			}
		} else {
			switch (card.CardSubtype) {
				case CardSubtype.MonsterNormal:
					s.backgroundImage = "url(card_frame/pendulum_normal.png)"
					break
				case CardSubtype.MonsterEffect:
					s.backgroundImage = "url(card_frame/pendulum_effect.png)"
					break
				case CardSubtype.MonsterRitual:
					s.backgroundImage = "url(card_frame/pendulum_ritual.png)"
					break
				case CardSubtype.MonsterFusion:
					s.backgroundImage = "url(card_frame/pendulum_fusion.png)"
					break
				case CardSubtype.MonsterSynchro:
					s.backgroundImage = "url(card_frame/pendulum_synchro.png)"
					break
				case CardSubtype.MonsterXyz:
					s.backgroundImage = "url(card_frame/pendulum_xyz.png)"
					break
				default:
					s.backgroundImage = "url(card_frame/pendulum_normal.png)"
					break
			}
		}
		// s.backgroundImage = "url(example/eg_monster_normal_blue_eyes.jpg)"
		// s.backgroundImage = "url(example/eg_monster_effect_flip_tuner.jpg)"
		// s.backgroundImage = "url(example/eg_monster_effect_long_text.jpg)"
		// s.backgroundImage = "url(example/eg_monster_fusion_non_effect_level12.jpg)"
		// s.backgroundImage = "url(example/eg_monster_synchro_baronne_real.jpg)"
		// s.backgroundImage = "url(example/eg_monster_synchro_level0.jpg)"
		// s.backgroundImage = "url(example/eg_monster_synchro_short_text.jpg)"
		// s.backgroundImage = "url(example/eg_monster_xyz_rank12.jpg)"
		// s.backgroundImage = "url(example/eg_monster_xyz_rank13.jpg)"
		// s.backgroundImage = "url(example/eg_monster_xyz_rank13.jpg)"
		// s.backgroundImage = "url(example/eg_monster_xyz_drident_real.jpg)"
		// s.backgroundImage = "url(example/eg_monster_link_5.jpg)"
		// s.backgroundImage = "url(example/eg_pendulum_effect_endymion.jpg)"
	}
}

function renderCardName(card) {
	let e = byId("RenderCardName")
	e.style.color = "black"
	if (card.CardType === CardType.Spell ||
		card.CardType === CardType.Trap) {
		e.style.color = "white"
	} else if (card.CardType === CardType.Monster) {
		if (card.CardSubtype === CardSubtype.MonsterXyz ||
			card.CardSubtype === CardSubtype.MonsterLink) {
			e.style.color = "white"
		}
	}
	fitTextOneLine(card.CardName, e, 1.5)
}

function renderCardAttribute(card) {
	let e = byId("ImgRenderCardAttribute")
	if (card.CardType === CardType.Spell) {
		e.src = MapImg.Spell
	} else if (card.CardType === CardType.Trap) {
		e.src = MapImg.Trap
	} else {
		e.src = MapImg[card.MonsterAttribute]
	}
}


function loadMonsterLevelRankElements() {
	{ // red star elements to represent monster level
		let mainElem = byId("RenderMonsterLevel")
		let nStars = 12
		let wrapStarStyle = function (style) {
			style.visibility = "hidden"
			style.display = "inline-block"
			style.width = Math.floor((mainElem.clientWidth) / nStars - 2) + "px"
			style.height = window.getComputedStyle(mainElem).height
			style.paddingLeft = "2px"
		}
		// star elements ID are StarWrap1, StarWrap2, ..., StarWrap12
		// they will be used in func renderCardTypeLevelRank
		for (let i = nStars; i >= 1; i--) {
			let starWrap = document.createElement("div")
			starWrap.id = `StarWrap${i}`
			wrapStarStyle(starWrap.style)
			let star = document.createElement("img")
			star.src = MapImg.Level
			star.style.width = "100%"
			starWrap.innerHTML = ''
			starWrap.appendChild(star)
			mainElem.appendChild(starWrap)
		}
	}
	{ // black star elements to represent monster rank (upto rank 12)
		let mainElem = byId("RenderMonsterRank")
		let nStars = 12
		let wrapStarStyle = function (style) {
			style.visibility = "hidden"
			style.display = "inline-block"
			style.width = Math.floor((mainElem.clientWidth) / nStars - 2) + "px"
			style.height = window.getComputedStyle(mainElem).height
			style.paddingRight = "2px"
		}
		// star elements ID are BlackStarWrap1, BlackStarWrap2, ..., BlackStarWrap12
		// they will be used in func renderCardTypeLevelRank
		for (let i = 1; i <= nStars; i++) {
			let starWrap = document.createElement("div")
			starWrap.id = `BlackStarWrap${i}`
			wrapStarStyle(starWrap.style)
			let star = document.createElement("img")
			star.src = MapImg.Rank
			star.style.width = "100%"
			starWrap.innerHTML = ''
			starWrap.appendChild(star)
			mainElem.appendChild(starWrap)
		}
	}
	{ // now YuGiOh only has 2 monsters that have rank 13:
		// * Raidraptor - Rising Rebellion Falcon
		// * Number iC1000: Numerounius Numerounia
		let mainElem = byId("RenderMonsterRank13")
		let nStars = 13
		let wrapStarStyle = function (style) {
			style.display = "inline-block"
			style.width = Math.floor((mainElem.clientWidth) / nStars - 1) + "px"
			style.height = window.getComputedStyle(mainElem).height
			style.paddingRight = "1px"
		}
		for (let i = 1; i <= nStars; i++) {
			let starWrap = document.createElement("div")
			wrapStarStyle(starWrap.style)
			let staticStar = document.createElement("img")
			staticStar.src = MapImg.Rank
			staticStar.style.width = "100%"
			starWrap.innerHTML = ''
			starWrap.appendChild(staticStar)
			mainElem.appendChild(starWrap)
		}
	}
}

function renderCardTypeLevelRank(card) {
	let level = byId("RenderMonsterLevel")
	let rank = byId("RenderMonsterRank")
	let rank13 = byId("RenderMonsterRank13")
	let cardType = byId("RenderCardType")
	let subType = byId("RenderCardSubtype")
	for (let v of [level, rank, rank13, cardType, subType]) {
		v.style.display = "none"
	}
	if (card.CardType === CardType.Monster) {
		if (card.CardSubtype === CardSubtype.MonsterLink) {
			return
		}
		if (card.CardSubtype !== CardSubtype.MonsterXyz) {
			level.style.display = ""
			for (let i = 1; i <= 12; i++) {
				if (i <= card.MonsterLevelRankLink) {
					byId(`StarWrap${i}`).style.visibility = "visible"
				} else {
					byId(`StarWrap${i}`).style.visibility = "hidden"
				}
			}
		} else if (card.MonsterLevelRankLink <= 12) {
			rank.style.display = ""
			for (let i = 1; i <= 12; i++) {
				if (i <= card.MonsterLevelRankLink) {
					byId(`BlackStarWrap${i}`).style.visibility = "visible"
				} else {
					byId(`BlackStarWrap${i}`).style.visibility = "hidden"
				}
			}
		} else {
			rank13.style.display = ""
		}
	} else { // Spell or Trap
		cardType.style.display = ""
		let textContent = ""
		if (card.CardType === CardType.Spell) {
			if (card.CardSubtype === CardSubtype.SpellNormal) {
				textContent = "[SPELL CARD]"
			} else {
				textContent = "[SPELL CARD  ]"
				subType.style.display = ""
				byId("ImgRenderCardSubtype").src = MapImg[card.CardSubtype]
			}
		} else if (card.CardType === CardType.Trap) {
			if (card.CardSubtype === CardSubtype.TrapNormal) {
				textContent = "[TRAP CARD]"
			} else {
				textContent = "[TRAP CARD  ]"
				subType.style.display = ""
				byId("ImgRenderCardSubtype").src = MapImg[card.CardSubtype]
			}
		}
		cardType.textContent = textContent
	}
}

let MapLinkMarker = {
	[LinkArrow.UpLeft]: "RenderLinkArrowUpLeft",
	[LinkArrow.Up]: "RenderLinkArrowUp",
	[LinkArrow.UpRight]: "RenderLinkArrowUpRight",
	[LinkArrow.Left]: "RenderLinkArrowLeft",
	[LinkArrow.Right]: "RenderLinkArrowRight",
	[LinkArrow.DownLeft]: "RenderLinkArrowDownLeft",
	[LinkArrow.Down]: "RenderLinkArrowDown",
	[LinkArrow.DownRight]: "RenderLinkArrowDownRight",
}

function renderLinkArrow(card) {
	for (let k in MapLinkMarker) {
		byId(MapLinkMarker[k]).style.visibility = "hidden"
	}
	if (card.CardSubtype !== CardSubtype.MonsterLink) {
		return
	}
	for (let v of card.MonsterLinkArrows) {
		let tmp = byId(MapLinkMarker[v])
		if (tmp) {tmp.style.visibility = "visible"}
	}
}

// renderCardEffect chooses and renders effect element based on card type and
// how long the effect is; this function also returns
// [chosenElement: HTMLElement, autoFontSize: number] for further processing
function renderCardEffect(card) {
	if (card.CardType === CardType.Spell || card.CardType === CardType.Trap) {
		let e = byId("RenderSpellTrapEffect")
		e.style.display = ""
		hideMonsterDetailElements()
		let innerHTML = card.CardEffect.replaceAll("\n", "<br>")
		e.innerHTML = innerHTML
		let s = window.getComputedStyle(e)
		let fontSizePx = chooseFontSize(innerHTML,
			e.clientWidth, e.clientHeight, s.fontFamily, s.fontWeight)
		e.style.fontSize = fontSizePx + "px"
		console.log(`chooseFontSize RenderSpellTrapEffect: ${fontSizePx}`)
		return [e, fontSizePx]
	}
	byId("RenderSpellTrapEffect").style.display = "none"

	// Monster effect and more detail:

	byId("RenderMonsterSplitLine").style.display = ""
	renderMonsterAtkDefLink(card)

	// choose RenderMonsterEffect or RenderMonsterEffectSmall

	let e1 = byId("RenderMonsterEffect")
	let a1 = byId("RenderMonsterAbilities")
	let s1 = window.getComputedStyle(e1)
	let e2 = byId("RenderMonsterEffectSmall")
	let a2 = byId("RenderMonsterAbilitiesSmall")
	let s2 = window.getComputedStyle(e2)
	let effectHTML = card.CardEffect.replaceAll("\n", "<br>")

	e1.style.display = ""
	a1.style.display = ""
	e2.style.display = "none"
	a2.style.display = "none"
	let fontSizePx = chooseFontSize(effectHTML,
		e1.clientWidth, e1.clientHeight, s1.fontFamily, s1.fontWeight)
	console.log(`chooseFontSize RenderMonsterEffect: ${fontSizePx}`)
	if (fontSizePx >= 34) {
		e1.style.fontSize = fontSizePx + "px"
		e1.innerHTML = effectHTML
		renderMonsterAbilities(card, a1)
		return [e1, fontSizePx]
	} else {
		e1.style.display = "none"
		a1.style.display = "none"
		e2.style.display = ""
		a2.style.display = ""
		let reFontSizePx = chooseFontSize(effectHTML,
			e2.clientWidth, e2.clientHeight, s2.fontFamily, s2.fontWeight)
		console.log(`chooseFontSize RenderMonsterEffectSmall: ${reFontSizePx}`)
		e2.style.fontSize = reFontSizePx + "px"
		e2.innerHTML = effectHTML
		renderMonsterAbilities(card, a2)
		return [e2, reFontSizePx]
	}
}

function textMonsterAbilities(card) {
	const separator = " / "
	let s = card.MonsterType

	if (card.CardSubtype === CardSubtype.MonsterRitual) {
		s += separator + "Ritual"
	} else if (card.CardSubtype === CardSubtype.MonsterFusion) {
		s += separator + "Fusion"
	} else if (card.CardSubtype === CardSubtype.MonsterSynchro) {
		s += separator + "Synchro"
	} else if (card.CardSubtype === CardSubtype.MonsterXyz) {
		s += separator + "Xyz"
	} else if (card.CardSubtype === CardSubtype.MonsterLink) {
		s += separator + "Link"
	}

	if (card.IsPendulum) {
		s += separator + "Pendulum"
	}
	if (card.MonsterAbilities) {
		for (let i = 0; i < card.MonsterAbilities.length; i++) {
			s += separator + card.MonsterAbilities[i]
		}
	}

	if (card.CardSubtype === CardSubtype.MonsterNormal) {
		s += separator + "Normal"
	} else if (card.CardSubtype === CardSubtype.MonsterEffect) {
		s += separator + "Effect"
	} else {
		if (card.CardEffect.trim().includes("\n")) {
			s += separator + "Effect"
		} else {
			// an extra deck non-effect monster or a ritual monster,
			// card text only has 1 line that is summoning condition.
		}
	}
	s = "[ " + s + " ]"
	return s
}

function renderMonsterAbilities(card, elementMonsterAbilities) {
	let tmp = textMonsterAbilities(card)
	fitTextOneLine(tmp, elementMonsterAbilities, 1.5)
}

function renderMonsterAtkDefLink(card) {
	if (card.CardType !== CardType.Monster) {
		return
	}
	let labelATK = byId("RenderMonsterATKLabel")
	let valueATK = byId("RenderMonsterATK")
	let labelDEF = byId("RenderMonsterDEFLabel")
	let valueDEF = byId("RenderMonsterDEF")
	let labelLINK = byId("RenderMonsterLinkLabel")
	let valueLINK = byId("RenderMonsterLinkRating")

	labelATK.style.display = ""
	valueATK.style.display = ""
	fitTextOneLine("ATK/", labelATK, 1.5, 1.15, 1.15)
	fitTextOneLine(card.MonsterATK, valueATK, 1.0, 1.25, 1.0)

	labelDEF.style.display = ""
	valueDEF.style.display = ""
	labelLINK.style.display = "none"
	valueLINK.style.display = "none"
	fitTextOneLine("DEF/", labelDEF, 1.5, 1.15, 1.15)
	fitTextOneLine(card.MonsterDEF, valueDEF, 1.0, 1.25, 1.0)

	// card.CardSubtype = CardSubtype.MonsterLink  // for testing
	if (card.CardSubtype === CardSubtype.MonsterLink) {
		labelDEF.style.display = "none"
		valueDEF.style.display = "none"
		labelLINK.style.display = ""
		valueLINK.style.display = ""
		fitTextOneLine("LINK-", labelLINK, 1.1, 0.96, 1.2)
		fitTextOneLine(card.MonsterLevelRankLink, valueLINK, 1, 1.0,)
	}
}

function renderMisc(card) {
	let kSet = byId("RenderKonamiSet")
	let kSetL = byId("RenderKonamiSetLink")
	let kSetP = byId("RenderKonamiSetPendulum")
	let kCid = byId("RenderKonamiCardID")
	let year = byId("RenderYearCreator")
	let all = [kSet, kSetP, kCid, year]
	if (card.IsPendulum) {
		kSet.style.display = "none"
		kSetL.style.display = "none"
		kSetP.style.display = ""
	} else if (card.CardSubtype === CardSubtype.MonsterLink) {
		kSet.style.display = "none"
		kSetL.style.display = ""
		kSetP.style.display = "none"
	} else {
		kSet.style.display = ""
		kSetL.style.display = "none"
		kSetP.style.display = "none"
	}
	if (card.CardSubtype === CardSubtype.MonsterXyz && !card.IsPendulum) {
		for (let v of all) {v.style.color = "white"}
	} else {
		for (let v of all) {v.style.color = "black"}
	}

	fitTextOneLine(card.MiscKonamiSet, kSet)
	fitTextOneLine(card.MiscKonamiSet, kSetL)
	fitTextOneLine(card.MiscKonamiSet, kSetP)
	fitTextOneLine(card.MiscKonamiCardID, kCid)
	fitTextOneLine(`🄯${card.MiscYear} ${card.MiscCreator}`, year)
}

function renderPendulum(card) {
	// console.log("renderPendulum", card.IsPendulum)
	let normalArt = byId("RenderCardArt")
	let pArt = byId("RenderCardArtPendulum")
	let pScaleL = byId("RenderPScaleLeft")
	let pScaleR = byId("RenderPScaleRight")
	let pEffect = byId("RenderPendulumEffect")
	if (!card.IsPendulum) {
		normalArt.style.display = ""
		pArt.style.display = "none"
		pScaleL.style.display = "none"
		pScaleR.style.display = "none"
		pEffect.style.display = "none"
		return
	}
	normalArt.style.display = "none"
	pArt.style.display = ""
	pScaleL.style.display = ""
	pScaleR.style.display = ""
	pEffect.style.display = ""
	fitTextOneLine(card.PendulumScale, pScaleL, 1.25)
	fitTextOneLine(card.PendulumScale, pScaleR, 1.25)
	let pEffectHTML = card.PendulumEffect.replaceAll("\n", "<br>")
	let s = window.getComputedStyle(pEffect)
	let fontSizePx = chooseFontSize(pEffectHTML,
		pEffect.clientWidth, pEffect.clientHeight, s.fontFamily, s.fontWeight)
	pEffect.style.fontSize = fontSizePx + "px"
	pEffect.innerHTML = pEffectHTML
}


function twitchFontCopyAuto() {
	byId("TwitchFont").value = byId("AutoFont").value
	byId("TwitchFontScaleY").value = "1.0"
}

function twitchFontCardEffect() {
	// cardEffectElement.id can be "RenderSpellTrapEffect" or
	// "RenderMonsterEffect" or
	// "RenderMonsterEffectSmall"
	let cardEffectElement = byId(byId("ChosenEffectElementID").value)
	if (!cardEffectElement) {
		return
	}
	let fontSize = document.getElementById("TwitchFont").value + "px"
	let scaleY = document.getElementById("TwitchFontScaleY").value
	cardEffectElement.style.fontSize = fontSize
	cardEffectElement.style.transform = `scale(1.0, ${scaleY})`
	cardEffectElement.style.transformOrigin = `top left`
}

// getText gets all text from an HTML element
function getText(node) {
	let resultArray = []

	function getTextRecur(node) {
		if (node.nodeType === 3) {// "3" is text node
			resultArray.push(node.nodeValue.trim())
		} else {
			for (let child of node.childNodes) {
				getTextRecur(child)
			}
		}
	}

	getTextRecur(node)
	return resultArray.join(" ")
}


function autoZoomPage() {
	let vpW = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0)
	let vpH = Math.max(document.documentElement.clientHeight || 0, window.innerHeight || 0)
	let docW = document.body.scrollWidth;
	let docH = document.body.scrollHeight;
	let time = (new Date()).toISOString()
	console.log("this app assumes display resolution are 3840x2160")
	console.log(`${time} view port: ${vpW}x${vpH}, document size: ${docW}x${docH}`)
	console.log("_____________________________________________________________")
	let scale = 1
	if (docW / vpW >= 1.5) {
		scale = 0.5
	}
	if (scale !== 1) {
		document.body.style.transform = `scale(${scale})`;
		document.body.style.transformOrigin = "0 0";
		// document.body.style["-moz-transform"] = `scale(${scale})`;
		// document.body.style["-moz-transform-origin"] = "0 0";
	}
}

// exportCardPNG downloads the rendered card HTML as a PNG image;
// https://stackoverflow.com/a/32776834/4097963:
// tried html2canvas, domtoimage,rasterizeHTML
function exportCardPNG() {
	let cardElem = byId("RenderCard")
	if (true) {
		html2canvas(cardElem).then(
			function (canvas) {
				downloadAsImage(canvas.toDataURL())
			})
	}

	if (true) {
		{
			let canvas = document.createElement("canvas");
			canvas.height = cardElem.clientHeight;
			canvas.width = cardElem.clientWidth;
			rasterizeHTML.drawHTML(cardElem.outerHTML, canvas)
				.then(function (renderResult) {
					downloadAsImage(canvas.toDataURL())
				});
		}
	}

	if (true) {
		domtoimage.toPng(cardElem, null).then(
			function (dataUrl) {
				console.log("hoho")
				downloadAsImage(dataUrl)
			})
			.catch(function (err) {
				console.log("haha")
				window.debug = err
				console.error(`error domtoimage: ${err}`);
			});
	}
}

// downloadAsImage makes browser download dataURL as a PNG image,
// output name based on current time
function downloadAsImage(dataURL) {
	let now = (new Date()).toISOString()
	now = now.replace(/[^A-Za-z0-9]/g, "");
	let outputFileName = `${now}.png`
	let link = document.createElement("a");
	link.id = "downloadAsImage"
	if (typeof link.download === "string") {
		link.href = dataURL;
		link.download = outputFileName;
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
	} else {
		window.open(dataURL);
	}
}

//  AllowChars are lowercase alphanumeric, good for file name cross-platform
let AllowChars = {}
for (let char of "abcdefghijklmnopqrstuvwxyz_0123456789".split("")) {
	AllowChars[char] = true
}

function normalizeFileName(str) {
	let ret = []
	for (let char of str.toLowerCase().split("")) {
		if (!AllowChars[char]) {
			ret.push('_')
		} else {
			ret.push(char)
		}
	}
	return ret.join("")
}

function exportCardJSON(isKeepCardArt = false) {
	console.log("begin function exportCardJSON()")
	if (!GlobalCard || !GlobalCard.CardName) {
		return
	}

	let card = CloneCard(GlobalCard)
	if (!isKeepCardArt) {
		card.CardArt = ""  // art file size is very big
	}
	let link = document.createElement("a");
	let beauty = JSON.stringify(card, null, "\t")
	link.href = `data:text/json;charset=utf-8,${encodeURIComponent(beauty)}`
	link.download = `${normalizeFileName(card.CardName)}_${card.MiscKonamiCardID}.json`
	document.body.appendChild(link);
	link.click();
	document.body.removeChild(link);
}

function importCardJSON(jsonDataURI) {
	// console.log(`importCardJSON jsonDataURI: ${jsonDataURI}`)
	// https://developer.mozilla.org/en-US/docs/Glossary/Base64#the_unicode_problem
	let prefixLen = "data:application/json;base64,".length
	let jsonStr = atob(jsonDataURI.substring(prefixLen))  // bad Unicode char
	let binStringUnicode = Uint8Array.from(jsonStr, (m) => m.codePointAt(0))
	let jsonStrUnicode = new TextDecoder().decode(binStringUnicode)
	console.log(`importCardJSON: ${jsonStrUnicode}`)
	GlobalCard = JSON.parse(jsonStrUnicode)
	loadCardToHTML(GlobalCard)
	renderCard(GlobalCard)
}


let CardDatabase = []
let MapCardDatabase = {}
let IndexCardDatabase  // lunr text search index

// https://github.com/olivernn/lunr.js
function buildIndexCardDatabase() {
	// build full text search index
	let beginT = new Date()
	// sleeping is a workaround that for the loop freeze the browser
	IndexCardDatabase = lunr(function () {
		this.field("CardName")
		// this.field("CardEffect")
		for (let i = 0; i < CardDatabase.length; i++) {
			let c = CardDatabase[i]
			this.add({
				"CardName": c.CardName,
				// "CardEffect": [c.CardEffect, c.PendulumEffect].join(" "),
				"id": c.MiscKonamiCardID,
			})
		}
		let duration = (new Date()) - beginT
		console.log(`indexed card database, dur: ${duration / 1000}s`)
	})
}

function konamiDatabaseURL(cardID, language = "ja") {
	return `https://www.db.yugioh-card.com/yugiohdb/card_search.action` +
		`?ope=2&request_locale=${language}&cid=${cardID}`
}

// https://github.com/olivernn/lunr.js
function searchCardDatabase() {
	let searchKey = document.getElementById("SearchCardQuery").value
	let matches = IndexCardDatabase.search(searchKey)
	if (!matches) {return}
	let limit = 10, offset = 0
	let result = matches.slice(offset, offset + limit)
	let resultWrap = document.getElementById("SearchCardResult")
	resultWrap.innerHTML = ""
	for (let v of result) {
		let card = JSON.parse(JSON.stringify(MapCardDatabase[v.ref]))
		if (!card) {continue}
		let row = document.createElement("div")
		row.className = "searchRow"
		let cardName = document.createElement("div")
		cardName.innerHTML = card.CardName
		row.appendChild(cardName)
		let summary = document.createElement("div")
		summary.style.textAlign = "right"
		if (card.CardType === CardType.Spell) {
			row.style.backgroundColor = "lightgreen"
			let text = MapCardSubtypeText[card.CardSubtype]
			summary.appendChild(document.createTextNode(text))
		} else if (card.CardType === CardType.Trap) {
			row.style.backgroundColor = "violet"
			let text = MapCardSubtypeText[card.CardSubtype]
			summary.appendChild(document.createTextNode(text))
		} else {
			row.style.backgroundColor = "khaki"
			let text = textMonsterAbilities(card)
			summary.appendChild(document.createTextNode(text))
		}
		for (let language of [/*"en",*/ "ja"]) {
			let konamiURL = document.createElement("a")
			konamiURL.href = konamiDatabaseURL(card.MiscKonamiCardID, language)
			konamiURL.target = "_blank"
			konamiURL.textContent = `  ${language}  `
			summary.appendChild(konamiURL)
		}
		row.appendChild(summary)
		row.onclick = function (mouseEvent) {
			if (mouseEvent.target !== this) { // click on a descendant
				if (mouseEvent.target.tagName === "A") {
					// open Konami link, do not refresh GlobalCard
					return
				}
			}
			// console.log(`mouseEvent searchRow: ${mouseEvent.target}`,)
			if (!card.MiscCreator) {  // keep "Creator"
				card.MiscCreator = byId("Creator").value
			}
			GlobalCard = card
			loadCardToHTML(GlobalCard)
			if (card.CardType === CardType.Monster) {
				byId("MonsterDetail").classList.remove("disabledElement")
			} else {
				byId("MonsterDetail").classList.add("disabledElement")
			}
			renderCard(GlobalCard)
		}
		resultWrap.appendChild(row)
	}
}


window.onload = () => {
	autoZoomPage()

	loadMonsterTypeElements()
	loadMonsterLevelRankElements()

	loadCardToHTML(DefaultCard)

	byId("CardArt").addEventListener("change", ev => {
		if (!ev.target.files || !ev.target.files.length) {return null}
		const r = new FileReader();
		r.onload = function (event) {
			if (window.getComputedStyle(byId("RenderCardArt")).display !== "none") {
				byId("ImgRenderCardArt").setAttribute("src", event.target.result)
			} else {
				byId("ImgRenderCardArtPendulum").setAttribute("src", event.target.result)
			}
			updateCardState()
		}
		r.readAsDataURL(ev.target.files[0]);
	});
	byId("ImportCardJSONFile").addEventListener("change", ev => {
		if (!ev.target.files || !ev.target.files.length) {return null}
		const r = new FileReader();
		r.onload = function (event) {
			// console.log(`data ImportCardJSONFile: ${event.target.result}`)
			importCardJSON(event.target.result)
		}
		r.readAsDataURL(ev.target.files[0]);
	});

	for (let v of document.getElementsByName("CardType")) {
		v.onclick = function () {handleClickCardType(v.id)}
	}

	byId("SearchCardQuery").addEventListener("keyup",
		function (event) {
			if (event.key === 'Enter') {
				byId("SearchCardDatabase").click()
			}
		});

	// almost all elements on "colLeft"
	let inputs = [
		document.getElementById("CardName"),
		// CardSubtypeWrap
		document.getElementById("MonsterNormal"),
		document.getElementById("MonsterEffect"),
		document.getElementById("MonsterRitual"),
		document.getElementById("MonsterFusion"),
		document.getElementById("MonsterSynchro"),
		document.getElementById("MonsterXyz"),
		document.getElementById("MonsterLink"),
		document.getElementById("SpellNormal"),
		document.getElementById("SpellQuickPlay"),
		document.getElementById("SpellRitual"),
		document.getElementById("SpellContinuous"),
		document.getElementById("SpellField"),
		document.getElementById("SpellEquip"),
		document.getElementById("TrapNormal"),
		document.getElementById("TrapCounter"),
		document.getElementById("TrapContinuous"),
		// Misc
		document.getElementById("Year"),
		document.getElementById("Creator"),
		document.getElementById("SetNumber"),
		document.getElementById("CardID"),
		//
		document.getElementById("CardEffect"),
		// MonsterAttributeWrap
		document.getElementById("DARK"),
		document.getElementById("EARTH"),
		document.getElementById("FIRE"),
		document.getElementById("LIGHT"),
		document.getElementById("WATER"),
		document.getElementById("WIND"),
		document.getElementById("DIVINE"),
		document.getElementById("MonsterType"),
		document.getElementById("MonsterLevelRankLink"),
		document.getElementById("MonsterATK"),
		document.getElementById("MonsterDEF"),
		// MonsterAbilitiesWrap
		document.getElementById("Flip"),
		document.getElementById("Gemini"),
		document.getElementById("Spirit"),
		document.getElementById("Toon"),
		document.getElementById("Tuner"),
		document.getElementById("Union"),
		// MonsterLinkArrows
		document.getElementById("UpLeft"),
		document.getElementById("Up"),
		document.getElementById("UpRight"),
		document.getElementById("Left"),
		document.getElementById("Right"),
		document.getElementById("DownLeft"),
		document.getElementById("Down"),
		document.getElementById("DownRight"),
		// PendulumWrap
		document.getElementById("IsPendulum"),
		document.getElementById("PendulumScale"),
		document.getElementById("PendulumEffect"),
	]
	for (let i = 0; i < inputs.length; i++) {
		inputs[i].oninput = handleInput
		inputs[i].onclick = handleClick
	}

	updateCardState()
	if (Boolean(window.chrome)) {
		// workaround Chromium based browsers calculate wrong font width
		// at the first load
		setTimeout(function () {renderCard(GlobalCard)}, 150)
	}


	let beginT = new Date()
	fetch('konami_data/konami_db_en.json').then(response => response.json()).then(
		function (data) {
			CardDatabase = data
			for (let v of data) {
				MapCardDatabase[v.MiscKonamiCardID] = v
			}
			let duration = (new Date()) - beginT
			console.log(`loaded card database, len: ${CardDatabase.length}, dur: ${duration / 1000}s`)

			buildIndexCardDatabase()
		}
	).catch(function (err) {console.log(`error cardDatabase: ${err}`)});
}
