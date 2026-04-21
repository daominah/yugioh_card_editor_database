package konami

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCardsToCSV(t *testing.T) {
	cards := []Card{
		{
			CardName:             "Blue-Eyes White Dragon",
			CardType:             "Monster",
			CardSubtype:          "MonsterNormal",
			CardEffect:           "This legendary dragon is a powerful engine of destruction. Virtually invincible, very few have faced this awesome creature and lived to tell the tale.",
			CardArt:              "",
			MonsterAttribute:     "LIGHT",
			MonsterType:          "Dragon",
			MonsterLevelRankLink: 8,
			MonsterATK:           3000,
			MonsterATKStr:        "3000",
			MonsterDEF:           2500,
			MonsterDEFStr:        "2500",
			MonsterAbilities:     nil,
			MonsterLinkArrows:    nil,
			IsNonEffectMonster:   true,
			IsPendulum:           false,
			PendulumScale:        0,
			PendulumEffect:       "",
			MiscKonamiSet:        "LOB-001",
			MiscKonamiCardID:     "4007",
			MiscCardPassword:     "89631139",
			MiscYear:             "2002",
			MiscCreator:          "",
		},
		{
			CardName:             "Maiden of White",
			CardType:             "Monster",
			CardSubtype:          "MonsterEffect",
			CardEffect:           "You can send this card from your hand or field to the GY; place 1 \"True Light\" from your hand, Deck, or GY, face-up in your Spell & Trap Zone. If you Special Summon \"Blue-Eyes White Dragon\" while this card is in your GY (except during the Damage Step): You can Special Summon this card. When a card or effect is activated that targets this card on the field, or when this card is targeted for an attack (Quick Effect): You can Special Summon 1 \"Blue-Eyes White Dragon\" or 1 Level 1 LIGHT Tuner from your GY. You can only use each effect of \"Maiden of White\" once per turn.",
			CardArt:              "",
			MonsterAttribute:     "LIGHT",
			MonsterType:          "Spellcaster",
			MonsterLevelRankLink: 1,
			MonsterATK:           0,
			MonsterATKStr:        "0",
			MonsterDEF:           0,
			MonsterDEFStr:        "0",
			MonsterAbilities:     []MonsterAbility{"Tuner"},
			MonsterLinkArrows:    nil,
			IsNonEffectMonster:   false,
			IsPendulum:           false,
			PendulumScale:        0,
			PendulumEffect:       "",
			MiscKonamiSet:        "SDWD-EN041",
			MiscKonamiCardID:     "20602",
			MiscCardPassword:     "17947697",
			MiscYear:             "2025",
			MiscCreator:          "",
		},
		{
			CardName:             "Blue-Eyes Spirit Dragon",
			CardType:             "Monster",
			CardSubtype:          "MonsterSynchro",
			CardEffect:           "1 Tuner + 1+ non-Tuner \"Blue-Eyes\" monsters\nNeither player can Special Summon 2 or more monsters at the same time. Once per turn, when an effect of a card in the GY is activated (Quick Effect): You can negate the activation. (Quick Effect): You can Tribute this Synchro Summoned card; Special Summon 1 LIGHT Dragon Synchro Monster from your Extra Deck in Defense Position, except \"Blue-Eyes Spirit Dragon\", but destroy it during the End Phase of this turn.",
			CardArt:              "",
			MonsterType:          "Dragon",
			MonsterLevelRankLink: 9,
			MonsterATK:           2500,
			MonsterATKStr:        "2500",
			MonsterDEF:           3000,
			MonsterDEFStr:        "3000",
			MonsterAbilities:     nil,
			MonsterLinkArrows:    nil,
			IsNonEffectMonster:   false,
			IsPendulum:           false,
			PendulumScale:        0,
			PendulumEffect:       "",
			MiscKonamiSet:        "SHVI-EN052",
			MiscKonamiCardID:     "12324",
			MiscCardPassword:     "59822133",
			MiscYear:             "2016",
			MiscCreator:          "",
		},
		{
			CardName:             "Wishes for Eyes of Blue",
			CardType:             "Spell",
			CardSubtype:          "SpellNormal",
			CardEffect:           "Discard 1 card; add 1 Level 1 LIGHT Tuner, and 1 Spell/Trap that mentions \"Blue-Eyes White Dragon\", from your Deck to your hand, except \"Wishes for Eyes of Blue\". You can banish this card from your GY, then target 1 \"Blue-Eyes White Dragon\" you control; equip 1 \"Blue-Eyes\" monster from your Extra Deck to it as an Equip Spell that gives it 400 ATK. You can only use each effect of \"Wishes for Eyes of Blue\" once per turn.",
			CardArt:              "",
			MonsterAttribute:     "",
			MonsterType:          "",
			MonsterLevelRankLink: 0,
			MonsterATK:           0,
			MonsterATKStr:        "",
			MonsterDEF:           0,
			MonsterDEFStr:        "",
			MonsterAbilities:     nil,
			MonsterLinkArrows:    nil,
			IsNonEffectMonster:   false,
			IsPendulum:           false,
			PendulumScale:        0,
			PendulumEffect:       "",
			MiscKonamiSet:        "SDWD-EN042",
			MiscKonamiCardID:     "20604",
			MiscCardPassword:     "80326401",
			MiscYear:             "2025",
			MiscCreator:          "",
		},
		{
			CardName:             "True Light",
			CardType:             "Trap",
			CardSubtype:          "TrapContinuous",
			CardArt:              "",
			MonsterAttribute:     "",
			MonsterType:          "",
			MonsterLevelRankLink: 0,
			MonsterATK:           0,
			MonsterATKStr:        "",
			MonsterDEF:           0,
			MonsterDEFStr:        "",
			MonsterAbilities:     nil,
			MonsterLinkArrows:    nil,
			IsNonEffectMonster:   false,
			IsPendulum:           false,
			PendulumScale:        0,
			PendulumEffect:       "",
			MiscKonamiSet:        "MP21-EN255",
			MiscKonamiCardID:     "16653",
			MiscCardPassword:     "62089826",
			MiscYear:             "2021",
			MiscCreator:          "",
		},
	}

	csvData := ToCSV(cards,
		map[string]string{
			"LOB":  "Legend of Blue Eyes White Dragon",
			"SDWD": "Structure Deck: Blue-Eyes White Destiny",
		})
	outputFile, err := os.Create("schema_test.csv")
	if err != nil {
		t.Fatalf("error creating CSV file: %v", err)
	}
	csvWriter := csv.NewWriter(outputFile)
	err = csvWriter.WriteAll(csvData)
	if err != nil {
		t.Fatalf("error csv.NewWriter.WriteAll: %v", err)
	}
	absPath, err := filepath.Abs(outputFile.Name())
	if err != nil {
		t.Fatalf("error getting absolute path: %v", err)
	}
	outputFile.Close()
	t.Logf("data written to file %v", absPath)

	// read the output file
	outputFile, err = os.Open(absPath)
	if err != nil {
		t.Fatalf("error os.Open: %v", err)
	}
	readData, err := io.ReadAll(outputFile)
	if err != nil {
		t.Fatalf("error io.ReadAll: %v", err)
	}
	want := `Maiden of White,Monster,Effect,20602,17947697,LIGHT,Spellcaster,1,0,0,Tuner,2025,SDWD-EN041,Structure Deck: Blue-Eyes White Destiny`
	if !strings.Contains(string(readData), want) {
		t.Errorf("bad output, got:\n%s, want contain:\n%s", string(readData), want)
	}
}
