package core

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/daominah/yugioh_card_editor_database/pkg/konami"
)

//go:embed ygocdb_card_password.json
var ygocdbData []byte // data is downloaded from https://ygocdb.com/api/v0/cards.zip

// CardPassword is the YGOCDB-derived password row keyed by Konami cardID.
// CardName is en_name with jp_name fallback so the row stays human-friendly
// when the card has not been released to TCG yet (no en_name).
type CardPassword struct {
	Password string
	CardName string
}

// MapCardsPasswordInitiator is an interface for initializing the map of
// Konami cardID to YGOCDB CardPassword (8-digit password and a name).
type MapCardsPasswordInitiator interface {
	InitMapCardsPassword() (map[konami.CardID]CardPassword, error)
}

// CardYgocdb data has addtional field Password (8-digit printed).
// Example data:
//
//	{
//	   "cid": 4007,
//	   "id": 89631139,
//	   "cn_name": "青眼白龙",
//	   "sc_name": "青眼白龙",
//	   "md_name": "青眼白龙",
//	   "nwbbs_n": "青眼白龙",
//	   "cnocg_n": "蓝眼白龙",
//	   "jp_ruby": "ブルーアイズ・ホワイト・ドラゴン",
//	   "jp_name": "青眼の白龍",
//	   "en_name": "Blue-Eyes White Dragon",
//	   "text": {
//	     "types": "[怪兽|通常] 龙/光\n[★8] 3000/2500",
//	     "pdesc": "",
//	     "desc": "以高攻击力著称的传说之龙。任何对手都能粉碎，其破坏力不可估量。"
//	   },
//	   "data": {
//	     "ot": 11,
//	     "setcode": 221,
//	     "type": 17,
//	     "atk": 3000,
//	     "def": 2500,
//	     "level": 8,
//	     "race": 8192,
//	     "attribute": 16
//	   }
//	 }
type CardYgocdb struct {
	Cid    int    `json:"cid"` // cid on Konami database
	Id     int    `json:"id"`  // Password
	CnName string `json:"cn_name"`
	ScName string `json:"sc_name"`
	MdName string `json:"md_name"`
	NwbbsN string `json:"nwbbs_n"`
	CnocgN string `json:"cnocg_n"`
	JpRuby string `json:"jp_ruby"`
	JpName string `json:"jp_name"`
	EnName string `json:"en_name"`
	Text   struct {
		Types string `json:"types"`
		Pdesc string `json:"pdesc"`
		Desc  string `json:"desc"`
	} `json:"text"`
	Data struct {
		Ot        int `json:"ot"`
		Setcode   int `json:"setcode"`
		Type      int `json:"type"`
		Atk       int `json:"atk"`
		Def       int `json:"def"`
		Level     int `json:"level"`
		Race      int `json:"race"`
		Attribute int `json:"attribute"`
	} `json:"data"`
}

// YgocdbStaticData implements MapCardsPasswordInitiator using embedded file static data.
type YgocdbStaticData struct{}

// InitMapCardsPassword returns the map of cardID to CardPassword from embedded static data.
func (s *YgocdbStaticData) InitMapCardsPassword() (map[konami.CardID]CardPassword, error) {
	return ParseYGOCDBData(ygocdbData)
}

// ParseYGOCDBData parses YGOCDB JSON data and returns a map of cardID to CardPassword.
// CardName is en_name when present, otherwise jp_name.
// This function is exported for use by driver implementations.
func ParseYGOCDBData(data []byte) (map[konami.CardID]CardPassword, error) {
	ygocdb := make(map[konami.CardID]CardYgocdb)
	err := json.Unmarshal(data, &ygocdb)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	m := make(map[konami.CardID]CardPassword)
	for _, v := range ygocdb {
		if v.Cid == 0 || v.Id == 0 {
			continue
		}
		name := v.EnName
		if name == "" {
			name = v.JpName
		}
		m[konami.CardID(strconv.Itoa(v.Cid))] = CardPassword{
			Password: fmt.Sprintf("%08d", v.Id), // pad with "0" to length 8
			CardName: name,
		}
	}
	if len(m) < 1000 { // expected about 14000 cards
		return nil, fmt.Errorf("parsed too few card passwords: %d", len(m))
	}
	return m, nil
}
