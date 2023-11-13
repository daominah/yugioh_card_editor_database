package core

// Card represent a YuGiOh card
type Card struct {
	CardName    string
	CardType    CardType
	CardSubtype CardSubtype
	CardEffect  string
	CardArt     string

	MonsterAttribute     MonsterAttribute
	MonsterType          MonsterType
	MonsterLevelRankLink int
	MonsterATK           float64
	MonsterATKStr        string // some monster have "?" ATK
	MonsterDEF           float64
	MonsterDEFStr        string // some monster have "?" DEF
	MonsterAbilities     []MonsterAbility
	MonsterLinkArrows    []MonsterLinkArrow

	// A non-Effect Monster is a Monster Card without a monster effect (this
	// includes all Normal Monsters (including Normal Pendulum monsters, token,
	// Gemini monsters) and certain Ritual, Fusion, Synchro, Xyz, and Link monster).
	IsNonEffectMonster bool

	IsPendulum     bool
	PendulumScale  int
	PendulumEffect string

	MiscKonamiSet    string
	MiscKonamiCardID string
	MiscYear         string
	MiscCreator      string
}

type CardType string
type CardSubtype string
type MonsterAttribute string
type MonsterType string
type MonsterAbility string
type MonsterLinkArrow string

const (
	Monster CardType = "Monster"
	Spell   CardType = "Spell"
	Trap    CardType = "Trap"

	MonsterNormal  CardSubtype = "MonsterNormal"
	MonsterEffect  CardSubtype = "MonsterEffect"
	MonsterRitual  CardSubtype = "MonsterRitual"
	MonsterFusion  CardSubtype = "MonsterFusion"
	MonsterSynchro CardSubtype = "MonsterSynchro"
	MonsterXyz     CardSubtype = "MonsterXyz"
	MonsterLink    CardSubtype = "MonsterLink"

	SpellNormal     CardSubtype = "SpellNormal"
	SpellQuickPlay  CardSubtype = "SpellQuickPlay"
	SpellRitual     CardSubtype = "SpellRitual"
	SpellContinuous CardSubtype = "SpellContinuous"
	SpellField      CardSubtype = "SpellField"
	SpellEquip      CardSubtype = "SpellEquip"

	TrapNormal     CardSubtype = "TrapNormal"
	TrapCounter    CardSubtype = "TrapCounter"
	TrapContinuous CardSubtype = "TrapContinuous"

	DARK   MonsterAttribute = "DARK"
	EARTH  MonsterAttribute = "EARTH"
	FIRE   MonsterAttribute = "FIRE"
	LIGHT  MonsterAttribute = "LIGHT"
	WATER  MonsterAttribute = "WATER"
	WIND   MonsterAttribute = "WIND"
	DIVINE MonsterAttribute = "DIVINE"

	Aqua         MonsterType = "Aqua"
	Beast        MonsterType = "Beast"
	BeastWarrior MonsterType = "Beast-Warrior"
	CreatorGod   MonsterType = "Creator God"
	Cyberse      MonsterType = "Cyberse"
	Dinosaur     MonsterType = "Dinosaur"
	DivineBeast  MonsterType = "Divine-Beast"
	Dragon       MonsterType = "Dragon"
	Fairy        MonsterType = "Fairy"
	Fiend        MonsterType = "Fiend"
	Fish         MonsterType = "Fish"
	Illusion     MonsterType = "Illusion" // example cardID 18812
	Insect       MonsterType = "Insect"
	Machine      MonsterType = "Machine"
	Plant        MonsterType = "Plant"
	Psychic      MonsterType = "Psychic"
	Pyro         MonsterType = "Pyro"
	Reptile      MonsterType = "Reptile"
	Rock         MonsterType = "Rock"
	SeaSerpent   MonsterType = "Sea Serpent"
	Spellcaster  MonsterType = "Spellcaster"
	Thunder      MonsterType = "Thunder"
	Warrior      MonsterType = "Warrior"
	WingedBeast  MonsterType = "Winged Beast"
	Wyrm         MonsterType = "Wyrm"
	Zombie       MonsterType = "Zombie"

	Flip   MonsterAbility = "Flip"
	Gemini MonsterAbility = "Gemini"
	Spirit MonsterAbility = "Spirit"
	Toon   MonsterAbility = "Toon"
	Tuner  MonsterAbility = "Tuner"
	Union  MonsterAbility = "Union"

	UpLeft    MonsterLinkArrow = "UpLeft"
	Up        MonsterLinkArrow = "Up"
	UpRight   MonsterLinkArrow = "UpRight"
	Left      MonsterLinkArrow = "Left"
	Right     MonsterLinkArrow = "Right"
	DownLeft  MonsterLinkArrow = "DownLeft"
	Down      MonsterLinkArrow = "Down"
	DownRight MonsterLinkArrow = "DownRight"
)
