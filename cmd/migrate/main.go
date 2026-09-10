// Command migrate applies or rolls back database migrations. The migration SQL
// is embedded in the binary (see the migrations package), so no external files
// or CLI are required.
//
// Usage:
//
//	go run ./cmd/migrate up          # apply all pending migrations
//	go run ./cmd/migrate down        # roll back the most recent migration
//	go run ./cmd/migrate down-all    # roll back every migration
//	go run ./cmd/migrate version     # print the current schema version
//	go run ./cmd/migrate force <v>   # clear a dirty state at version <v>
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/SagarR674/todo-api/config"
	"github.com/SagarR674/todo-api/database"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up|down|down-all|version|force <v>>")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	switch os.Args[1] {
	case "up":
		must(database.Migrate(cfg, database.Up))
		fmt.Println("migrations applied")
	case "down":
		must(database.Migrate(cfg, database.StepDown))
		fmt.Println("rolled back one migration")
	case "down-all":
		must(database.Migrate(cfg, database.Down))
		fmt.Println("all migrations rolled back")
	case "version":
		v, dirty, err := database.MigrationVersion(cfg)
		must(err)
		fmt.Printf("version=%d dirty=%t\n", v, dirty)
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("usage: migrate force <version>")
		}
		v, err := strconv.Atoi(os.Args[2])
		must(err)
		must(database.MigrationForce(cfg, v))
		fmt.Printf("forced to version %d\n", v)
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
}

func must(err error) {
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}
