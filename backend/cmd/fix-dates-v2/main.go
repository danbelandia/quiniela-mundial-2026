package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

type matchEntry struct {
	ID       int    `json:"id"`
	FechaUTC string `json:"fecha_utc"`
	Equipo1  string `json:"equipo1"`
	Equipo2  string `json:"equipo2"`
	Grupo    string `json:"grupo"`
}

var teamNameMap = map[string]string{
	"República de Corea": "Corea del Sur",
	"Arabia Saudí":       "Arabia Saudita",
	"RI de Irán":         "Irán",
	"RD Congo":           "República Democrática del Congo",
	"Sudam\u00e9rica":    "Sud\u00e1frica",
}

func translate(name string) string {
	if mapped, ok := teamNameMap[name]; ok {
		return mapped
	}
	return name
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
	exe, _ := os.Executable()
	dbPath := filepath.Join(filepath.Dir(exe), "..", "..", "..", "quiniela.db")
	if _, err := os.Stat(dbPath); err == nil {
		log.Printf("Usando SQLite local en %s", dbPath)
		return "sqlite", dbPath
	}
	log.Println("quiniela.db no encontrado junto al binario, usando 'quiniela.db' (cwd)")
	return "sqlite", "quiniela.db"
}

func main() {
	exe, _ := os.Executable()
	jsonPath := filepath.Join(filepath.Dir(exe), "matches.json")
	if _, err := os.Stat(jsonPath); err != nil {
		jsonPath = "matches.json"
	}
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatalf("read %s: %v", jsonPath, err)
	}
	var entries []matchEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		log.Fatalf("parse matches.json: %v", err)
	}
	log.Printf("Loaded %d match entries from matches.json", len(entries))

	driverName, connStr := buildDBConfig()
	db, err := sql.Open(driverName, connStr)
	if err != nil {
		log.Fatalf("open: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	updated, skipped, notFound := 0, 0, 0
	for _, e := range entries {
		if _, err := time.Parse(time.RFC3339, e.FechaUTC); err != nil {
			log.Fatalf("invalid fecha_utc for id=%d: %v", e.ID, err)
		}
		t1 := translate(e.Equipo1)
		t2 := translate(e.Equipo2)

		var currentID int
		err := db.QueryRow(`
			SELECT id FROM matches
			WHERE group_name = ?
			  AND ((home_team = ? AND away_team = ?) OR (home_team = ? AND away_team = ?))
			LIMIT 1
		`, e.Grupo, t1, t2, t2, t1).Scan(&currentID)

		if err == sql.ErrNoRows {
			log.Printf("NOT FOUND: group=%s %s vs %s (new_id=%d, date=%s)", e.Grupo, t1, t2, e.ID, e.FechaUTC)
			notFound++
			continue
		}
		if err != nil {
			log.Fatalf("query group=%s %s vs %s: %v", e.Grupo, t1, t2, err)
		}

		var oldDate string
		if err := db.QueryRow("SELECT match_date FROM matches WHERE id = ?", currentID).Scan(&oldDate); err != nil {
			log.Fatalf("read current date for id=%d: %v", currentID, err)
		}
		if oldDate == e.FechaUTC {
			skipped++
			continue
		}

		res, err := db.Exec("UPDATE matches SET match_date = ? WHERE id = ?", e.FechaUTC, currentID)
		if err != nil {
			log.Fatalf("update id=%d: %v", currentID, err)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			log.Printf("update id=%d: 0 rows affected", currentID)
			continue
		}
		log.Printf("OK id=%d (%s vs %s, group %s): %s -> %s", currentID, t1, t2, e.Grupo, oldDate, e.FechaUTC)
		updated++
	}

	fmt.Printf("\nSummary: %d updated, %d unchanged, %d not-found (out of %d)\n", updated, skipped, notFound, len(entries))
}
