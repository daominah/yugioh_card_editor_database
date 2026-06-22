// aggregate-type-attr-rarity is the CLI entry point for core.Aggregate.
// It opens the project SQLite DB and delegates orchestration of the five
// lookup-table aggregation passes (monster_attributes, monster_types, *_rush,
// rarities) to pkg/core via the core.DatabaseAggregate interface, which
// pkg/driver/sqlite implements.
package main

import (
	"log"
	"path/filepath"

	"github.com/daominah/yugioh_card_editor/pkg/base"
	"github.com/daominah/yugioh_card_editor/pkg/core"
	"github.com/daominah/yugioh_card_editor/pkg/driver/sqlite"
)

func main() {
	projectRoot, err := base.GetProjectRootDir()
	if err != nil {
		log.Fatalf("error base.GetProjectRootDir: %v", err)
	}
	db, err := sqlite.Open(filepath.Join(projectRoot, "data/yugioh.db"))
	if err != nil {
		log.Fatalf("error sqlite.Open: %v", err)
	}
	defer db.Close()

	if err := core.Aggregate(db); err != nil {
		log.Fatalf("error core.Aggregate: %v", err)
	}
}
