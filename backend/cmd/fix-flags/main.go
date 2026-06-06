package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

func main() {
	driverName, connStr := buildDBConfig()

	db, err := sql.Open(driverName, connStr)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	flags := map[string]string{
		"Escocia":    "🏴󠁧󠁢󠁳󠁣󠁴󠁿",
		"Inglaterra": "🏴󠁧󠁢󠁥󠁮󠁧󠁿",
	}

	total := 0
	for team, flag := range flags {
		res, err := db.Exec("UPDATE matches SET home_flag = ? WHERE home_team = ?", flag, team)
		if err != nil {
			log.Fatalf("update home %s: %v", team, err)
		}
		n, _ := res.RowsAffected()
		log.Printf("home %s: %d row(s)", team, n)
		total += int(n)

		res, err = db.Exec("UPDATE matches SET away_flag = ? WHERE away_team = ?", flag, team)
		if err != nil {
			log.Fatalf("update away %s: %v", team, err)
		}
		n, _ = res.RowsAffected()
		log.Printf("away %s: %d row(s)", team, n)
		total += int(n)
	}

	fmt.Printf("Total rows updated: %d\n", total)
}

func buildDBConfig() (string, string) {
	url := os.Getenv("TURSO_DATABASE_URL")
	token := os.Getenv("TURSO_AUTH_TOKEN")

	if url != "" {
		if token == "" {
			log.Fatal("TURSO_DATABASE_URL está definida pero falta TURSO_AUTH_TOKEN")
		}
		log.Println("Conectando a Turso...")
		return "libsql", fmt.Sprintf("%s?authToken=%s", url, token)
	}

	log.Println("TURSO_DATABASE_URL no definida, usando SQLite local (quiniela.db)")
	return "sqlite", "quiniela.db"
}
