package konami

import (
	"slices"
	"strconv"
	"strings"
)

// KonamiDB identifies which Konami web database to query.
type KonamiDB string

const (
	StandardDB KonamiDB = "yugiohdb" // TCG / OCG / Master Duel
	RushDB     KonamiDB = "rushdb"   // Rush Duel / Duel Links
)

// Card represent a YuGiOh card
type Card struct {
	CardName string
	// CardNameEN is the English card name as printed on the Japanese card
	// page <h1> (the bare <span> next to the kanji title). It is parsed and
	// written only during the "ja" crawl pass; the EN crawl page itself does
	// not source it. Stored on the cards row as a viewer convenience so the
	// DB can be browsed without joining card_texts. The authoritative
	// localized EN name still lives in card_texts where lang='en'.
	// Always empty for Rush Duel cards (Konami's Rush JA pages have no EN span).
	CardNameEN  string
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

	// IsSpecialSummonOnly marks "Nomi" / "Semi-Nomi" monsters that cannot be
	// Normal Summoned or Set; they can only be Special Summoned via the
	// conditions printed on the card text. Detected from the JA token "特殊召喚"
	// (or KO "특수 소환") in the species line. Konami's EN page omits this flag
	// from the species block, so the value is sourced from the JA parse pass.
	IsSpecialSummonOnly bool

	IsPendulum     bool
	PendulumScale  int
	PendulumEffect string

	MiscKonamiSet string // a.k.a. "Card Number", e.g. "LB-01", "LOB-001", "LOB-EN001"
	// MiscKonamiCardID is the card's numeric ID in Konami's online database,
	// taken from the cid= URL parameter, e.g. "Blue-Eyes White Dragon" has 4007.
	// The same numeric ID identifies the card across Standard (TCG/OCG) and is
	// disjoint from Rush Duel IDs, so it is used as the INTEGER primary key in
	// the cards / cards_rush tables (see init_schema.sql).
	// Although CardID is a string type for ergonomics at parse boundaries,
	// any ordering or range comparison must go through CardID.Int() — the IDs
	// are numeric, and lexicographic string compare gives the wrong order
	// (e.g. "10000" < "4007" as strings).
	MiscKonamiCardID CardID
	// 8-digit Password printed on the bottom left of a card,
	// in the past was used to unlock cards in video games,
	// Konami database does not show this information.
	MiscCardPassword string // e.g. "89631139"
	MiscYear         string // the year the card was released in the TCG (usually after the Japanese release)
	MiscCreator      string
}

type (
	CardID           string
	CardType         string
	CardSubtype      string
	MonsterAttribute string
	MonsterType      string
	MonsterAbility   string
	MonsterLinkArrow string
)

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
)

const (
	DARK   MonsterAttribute = "DARK"
	EARTH  MonsterAttribute = "EARTH"
	FIRE   MonsterAttribute = "FIRE"
	LIGHT  MonsterAttribute = "LIGHT"
	WATER  MonsterAttribute = "WATER"
	WIND   MonsterAttribute = "WIND"
	DIVINE MonsterAttribute = "DIVINE"

	Flip        MonsterAbility = "Flip"
	Gemini      MonsterAbility = "Gemini"
	Spirit      MonsterAbility = "Spirit"
	Toon        MonsterAbility = "Toon"
	Tuner       MonsterAbility = "Tuner"
	Union       MonsterAbility = "Union"
	RushMaximum MonsterAbility = "Maximum"

	UpLeft    MonsterLinkArrow = "UpLeft"
	Up        MonsterLinkArrow = "Up"
	UpRight   MonsterLinkArrow = "UpRight"
	Left      MonsterLinkArrow = "Left"
	Right     MonsterLinkArrow = "Right"
	DownLeft  MonsterLinkArrow = "DownLeft"
	Down      MonsterLinkArrow = "Down"
	DownRight MonsterLinkArrow = "DownRight"
)

// OCG / TCG / Master Duel 26 monster types
const (
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
)

// Rush Duel has 29 monster types:
// - 23 shared with OCG/TCG (all except Creator God, Divine-Beast, Illusion)
// - 6 unique to Rush Duel (constants below); Rush Duel is only available in JA/KO so EN names are unofficial
const (
	MagicalKnight    MonsterType = "Magical Knight"
	Cyborg           MonsterType = "Cyborg"
	HighDragon       MonsterType = "High Dragon"
	CelestialWarrior MonsterType = "Celestial Warrior"
	OmegaPsychic     MonsterType = "Omega Psychic"
	Galaxy           MonsterType = "Galaxy"
)

// CardCounts holds per-dimension card counts from the cards table.
// Monster-specific dimensions (MonsterAttribute, MonsterType, Level, ATK, DEF)
// only include rows where card_type = 'Monster'.
type CardCounts struct {
	ByType             map[CardType]int
	BySubtype          map[CardSubtype]int
	ByMonsterAttribute map[MonsterAttribute]int
	ByMonsterType      map[MonsterType]int
	ByMonsterLevel     map[int]int
	ByMonsterATK       map[int]int
	ByMonsterDEF       map[int]int
}

// Int convert CardID (string) to int, return 0 if error
func (id CardID) Int() int {
	ret, err := strconv.Atoi(string(id))
	if err != nil {
		return 0
	}
	return ret
}

type SortCardIDs []CardID

func (s SortCardIDs) Len() int           { return len(s) }
func (s SortCardIDs) Less(i, j int) bool { return s[i].Int() < s[j].Int() }
func (s SortCardIDs) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type SortCardNames []Card

func (s SortCardNames) Len() int           { return len(s) }
func (s SortCardNames) Less(i, j int) bool { return s[i].CardName < s[j].CardName }
func (s SortCardNames) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// ToCSV output can be used for csv.Writer.WriteAll
func ToCSV(cards []Card, mapSetsFullName map[string]string) [][]string {
	outputFields := []string{
		"Row",
		"ENName", "CardType", "CardSubtype",
		"CardID", "CardPasswd",
		"Attribute", "Type", "Level", "ATK", "DEF", "Tuner",
		"ENYear", "ENSet", "ENSetFullName",
	}
	records := [][]string{outputFields}
	for i, c := range cards {
		isTuner := slices.Contains(c.MonsterAbilities, Tuner)
		var isTunerStr string
		if isTuner {
			isTunerStr = "Tuner"
		}

		atkStr := c.MonsterATKStr
		if atkStr == "" && c.CardType == Monster {
			atkStr = strconv.Itoa(int(c.MonsterATK))
		}
		defStr := c.MonsterDEFStr
		if defStr == "" && c.CardType == Monster {
			defStr = strconv.Itoa(int(c.MonsterDEF))
		}
		levelStr := strconv.Itoa(c.MonsterLevelRankLink)
		if c.CardType != Monster {
			levelStr = ""
		}

		cardNumber := c.MiscKonamiSet // "LOB-001"
		firstDash := strings.Index(cardNumber, "-")
		if firstDash > -1 {
			cardNumber = cardNumber[:firstDash] // "LOB"
		}
		setFullName := mapSetsFullName[cardNumber]
		record := []string{
			strconv.Itoa(i + 1), // so we can see the row count in PDF

			c.CardName,
			string(c.CardType),
			strings.TrimPrefix(string(c.CardSubtype), string(c.CardType)),

			string(c.MiscKonamiCardID),
			c.MiscCardPassword,

			string(c.MonsterAttribute),
			string(c.MonsterType),
			levelStr,
			atkStr,
			defStr,
			isTunerStr,

			c.MiscYear,
			c.MiscKonamiSet,
			setFullName,
		}
		records = append(records, record)
	}
	return records
}
