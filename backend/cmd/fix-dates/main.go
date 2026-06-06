package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

var matchDates = []struct {
	MatchID int
	Date    string
}{
	{1, "2026-06-11T19:00:00Z"}, {2, "2026-06-12T02:00:00Z"}, {3, "2026-06-12T19:00:00Z"},
	{4, "2026-06-13T03:00:00Z"}, {5, "2026-06-13T19:00:00Z"}, {6, "2026-06-13T21:00:00Z"},
	{7, "2026-06-14T00:00:00Z"}, {8, "2026-06-14T03:00:00Z"}, {9, "2026-06-14T17:00:00Z"},
	{10, "2026-06-14T20:00:00Z"}, {11, "2026-06-14T23:00:00Z"}, {12, "2026-06-15T02:00:00Z"},
	{13, "2026-06-15T16:00:00Z"}, {14, "2026-06-15T19:00:00Z"}, {15, "2026-06-15T22:00:00Z"},
	{16, "2026-06-16T01:00:00Z"}, {17, "2026-06-16T19:00:00Z"}, {18, "2026-06-16T22:00:00Z"},
	{19, "2026-06-17T01:00:00Z"}, {20, "2026-06-17T04:00:00Z"}, {21, "2026-06-17T17:00:00Z"},
	{22, "2026-06-17T20:00:00Z"}, {23, "2026-06-17T23:00:00Z"}, {24, "2026-06-18T02:00:00Z"},
	{25, "2026-06-18T16:00:00Z"}, {26, "2026-06-18T19:00:00Z"}, {27, "2026-06-18T22:00:00Z"},
	{28, "2026-06-19T01:00:00Z"}, {29, "2026-06-19T19:00:00Z"}, {30, "2026-06-19T22:00:00Z"},
	{31, "2026-06-20T01:00:00Z"}, {32, "2026-06-20T04:00:00Z"}, {33, "2026-06-20T17:00:00Z"},
	{34, "2026-06-20T20:00:00Z"}, {35, "2026-06-21T02:00:00Z"}, {36, "2026-06-21T04:00:00Z"},
	{37, "2026-06-21T16:00:00Z"}, {38, "2026-06-21T19:00:00Z"}, {39, "2026-06-21T22:00:00Z"},
	{40, "2026-06-22T01:00:00Z"}, {41, "2026-06-22T17:00:00Z"}, {42, "2026-06-22T21:00:00Z"},
	{43, "2026-06-23T00:00:00Z"}, {44, "2026-06-23T03:00:00Z"}, {45, "2026-06-23T17:00:00Z"},
	{46, "2026-06-23T20:00:00Z"}, {47, "2026-06-23T23:00:00Z"}, {48, "2026-06-24T02:00:00Z"},
	{49, "2026-06-24T19:00:00Z"}, {50, "2026-06-24T19:00:00Z"}, {51, "2026-06-24T22:00:00Z"},
	{52, "2026-06-24T22:00:00Z"}, {53, "2026-06-25T01:00:00Z"}, {54, "2026-06-25T01:00:00Z"},
	{55, "2026-06-25T20:00:00Z"}, {56, "2026-06-25T20:00:00Z"}, {57, "2026-06-25T23:00:00Z"},
	{58, "2026-06-25T23:00:00Z"}, {59, "2026-06-26T02:00:00Z"}, {60, "2026-06-26T02:00:00Z"},
	{61, "2026-06-26T19:00:00Z"}, {62, "2026-06-26T19:00:00Z"}, {63, "2026-06-27T00:00:00Z"},
	{64, "2026-06-27T00:00:00Z"}, {65, "2026-06-27T03:00:00Z"}, {66, "2026-06-27T03:00:00Z"},
	{67, "2026-06-27T21:00:00Z"}, {68, "2026-06-27T21:00:00Z"}, {69, "2026-06-27T23:30:00Z"},
	{70, "2026-06-27T23:30:00Z"}, {71, "2026-06-28T02:00:00Z"}, {72, "2026-06-28T02:00:00Z"},
}

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

	total := 0
	for _, md := range matchDates {
		_, err := time.Parse(time.RFC3339, md.Date)
		if err != nil {
			log.Fatalf("invalid date for match_id %d: %v", md.MatchID, err)
		}
		res, err := db.Exec("UPDATE matches SET match_date = ? WHERE id = ?", md.Date, md.MatchID)
		if err != nil {
			log.Fatalf("update match %d: %v", md.MatchID, err)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			log.Printf("match_id %d: 0 rows (id not found?)", md.MatchID)
		}
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
