package konami

// RarityCode is the short Konami identifier for a card printing rarity,
// matching the rarity_code column in the rarities and set_cards tables.
type RarityCode string

// RarityCode enum values are created from the aggregated SQLite table "rarities"
// (not used anywhere in logic code yet)
const (
	// Standard rarities in Basic Packs:

	N  RarityCode = "N"  // Normal (JA/KO); equivalent of C
	C  RarityCode = "C"  // Common (JA/EN); equivalent of N
	R  RarityCode = "R"  // Rare
	SR RarityCode = "SR" // Super Rare
	UR RarityCode = "UR" // Ultra Rare
	SE RarityCode = "SE" // Secret Rare

	// Embossing rarities:

	UL RarityCode = "UL" // Ultimate Rare
	CR RarityCode = "CR" // Collector's Rare

	HR RarityCode = "HR" // Holographic Rare (JA/KO)
	GH RarityCode = "GH" // Ghost Rare (EN)

	// fully flashy card

	PSE  RarityCode = "PSE"  // Prismatic Secret Rare; equivalent of STAR
	STAR RarityCode = "STAR" // Starlight Rare (EN); equivalent of PSE

	EXSE RarityCode = "EXSE" // Extra Secret Rare (JA/KO), equivalent of Platinum Secret Rare (EN)
	PS   RarityCode = "PS"   // Platinum Secret Rare (EN); equivalent of EXSE

	SE20th  RarityCode = "20th SE"  // 20th Anniversary Secret Rare (JA)
	SE10000 RarityCode = "10000 SE" // 10000 Secret Rare (the same foil patterns as a 20th SE)
	QCSE    RarityCode = "QCSE"     // Quarter Century Secret Rare (25th SE)

	// GMR is Grand Master Rare (JA), exclusive to Over-Frame
	GMR RarityCode = "GMR"

	// Gold variants:

	GR  RarityCode = "GR"  // Gold Rare
	GSE RarityCode = "GSE" // Gold Secret Rare
	PGR RarityCode = "PGR" // Premium Gold Rare (EN); equivalent of PG
	PG  RarityCode = "PG"  // Premium Gold Rare (JA/KO); equivalent of PGR

	// Parallel variant:

	P      RarityCode = "P"      // (JA/KO)
	P_UR   RarityCode = "P+UR"   // (JA/KO)
	P_SR   RarityCode = "P+SR"   // (JA/KO)
	P_SE   RarityCode = "P+SE"   // (JA/KO)
	P_ES   RarityCode = "P+ES"   // KO parallel Extra Secret variant
	P_EXSE RarityCode = "P+EXSE" // (JA)
	P_HR   RarityCode = "P+HR"   // (JA)
	P_R    RarityCode = "P+R"    // (JA)

	// Millennium variant (JA/KO):

	M    RarityCode = "M"
	M_SR RarityCode = "M+SR"
	M_UR RarityCode = "M+UR"
	M_GR RarityCode = "M+GR"
	M_SE RarityCode = "M+SE"

	UR_PR RarityCode = "UR (PR)" // Pharaoh's Rare promo (EN); equivalent of Millennium Rare

	// KC variant (JA):

	KC_UR RarityCode = "KC+UR"
	KC    RarityCode = "KC"
	KC_R  RarityCode = "KC+R"

	// Rush Duel rarities:

	RR  RarityCode = "RR"  // Rush Rare (JA/KO)
	GRR RarityCode = "GRR" // Gold Rush Rare (JA)
	ORR RarityCode = "ORR" // Over Rush Rare (JA/KO)
	FOR RarityCode = "FOR" // Full Over Rush Rare (JA)

	// Deprecated rarities (no longer printed):

	UR_Hobby RarityCode = "UR (Hobby)" // Hobby League promo (EN); early 2000s
	H        RarityCode = "H"          // Hobby (EN); DB label for Battle Pack foil rarities (2012–2014)
	ST       RarityCode = "ST"         // Starfoil (EN); Star Packs / Battle Pack, last printed 2014
	SH       RarityCode = "SH"         // Shatterfoil (EN); Battle Pack 3, only printed 2014
	MR       RarityCode = "MR"         // Mosaic Rare (EN); Battle Pack 2, last printed 2014
	PL       RarityCode = "PL"         // Platinum Rare (EN); Noble Knights of the Round Table, only printed 2014
)

// Name returns the English display name for the rarity code
// (some are shortened and not canonical)
func (rarityCode RarityCode) Name() string {
	switch rarityCode {
	case N:
		return "Normal"
	case C:
		return "Common"
	case R:
		return "Rare"
	case SR:
		return "Super Rare"
	case UR:
		return "Ultra Rare"
	case SE:
		return "Secret Rare"
	case UL:
		return "Ultimate Rare"
	case CR:
		return "Collector Rare"
	case HR:
		return "Holographic Rare"
	case GH:
		return "Ghost Rare"
	case PSE:
		return "Prismatic Secret"
	case STAR:
		return "Starlight Rare"
	case EXSE:
		return "Extra Secret"
	case PS:
		return "Platinum Secret"
	case SE20th:
		return "20th Secret"
	case SE10000:
		return "10000 Secret"
	case QCSE:
		return "Quarter Century Secret"
	case GMR:
		return "Grand Master"
	case GR:
		return "Gold Rare"
	case GSE:
		return "Gold Secret"
	case PGR:
		return "Premium Gold"
	case PG:
		return "Premium Gold"
	case P:
		return "Parallel"
	case P_R:
		return "Parallel Rare"
	case P_SR:
		return "Parallel Super"
	case P_UR:
		return "Parallel Ultra"
	case P_SE:
		return "Parallel Secret"
	case P_ES:
		return "Parallel Extra Secret"
	case P_EXSE:
		return "Parallel Extra Secret"
	case P_HR:
		return "Parallel Holographic"
	case M:
		return "Millennium Rare"
	case M_SR:
		return "Millennium Super"
	case M_UR:
		return "Millennium Ultra"
	case M_GR:
		return "Millennium Gold"
	case M_SE:
		return "Millennium Secret"
	case UR_PR:
		return "Ultra Rare Pharaoh"
	case KC_UR:
		return "Kaiba Corp Ultra"
	case KC:
		return "Kaiba Corp"
	case KC_R:
		return "Kaiba Corp"
	case RR:
		return "Rush Rare"
	case GRR:
		return "Gold Rush"
	case ORR:
		return "Over Rush"
	case FOR:
		return "Full Over Rush"
	case UR_Hobby:
		return "Ultra Rare Hobby League"
	case H:
		return "Hobby Rare"
	case ST:
		return "Starfoil Rare"
	case SH:
		return "Shatterfoil Rare"
	case MR:
		return "Mosaic Rare"
	case PL:
		return "Platinum Rare"
	}
	return string(rarityCode)
}
