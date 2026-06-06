package repository

import (
	"database/sql"
	"log"
	"quiniela-backend/internal/core/domain"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	DB *sql.DB
}

func NewSQLiteRepository(driverName, connStr string) (*SQLiteRepository, error) {
	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	repo := &SQLiteRepository{DB: db}
	if err := repo.Migrate(); err != nil {
		return nil, err
	}
	if err := repo.Seed(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteRepository) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		is_admin BOOLEAN DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS matches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		home_team TEXT NOT NULL,
		away_team TEXT NOT NULL,
		home_score INTEGER DEFAULT 0,
		away_score INTEGER DEFAULT 0,
		match_date DATETIME NOT NULL,
		status TEXT NOT NULL,
		group_name TEXT NOT NULL,
		is_locked BOOLEAN DEFAULT 0,
		home_flag TEXT DEFAULT '',
		away_flag TEXT DEFAULT ''
	);
	CREATE TABLE IF NOT EXISTS predictions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		match_id INTEGER NOT NULL,
		home_score INTEGER NOT NULL,
		away_score INTEGER NOT NULL,
		UNIQUE(user_id, match_id),
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(match_id) REFERENCES matches(id)
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users(email);`
	_, err := r.DB.Exec(query)
	return err
}

func (r *SQLiteRepository) Seed() error {
	var count int
	r.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count == 0 {
		_, err := r.DB.Exec("INSERT INTO users (username, email, password, is_admin) VALUES (?, ?, ?, ?)",
			"dangel", "admin@quiniela.com", "281193", 1)
		if err != nil {
			return err
		}
	} else {
		_, err := r.DB.Exec("UPDATE users SET is_admin = 1, password = ? WHERE username = 'dangel'", "281193")
		if err != nil {
			return err
		}
	}

	r.DB.QueryRow("SELECT COUNT(*) FROM matches").Scan(&count)
	if count > 0 {
		return nil
	}

	flags := map[string]string{
		"México": "🇲🇽", "Sudáfrica": "🇿🇦", "Corea del Sur": "🇰🇷", "República Checa": "🇨🇿",
		"Canadá": "🇨🇦", "Bosnia y Herzegovina": "🇧🇦", "Catar": "🇶🇦", "Suiza": "🇨🇭",
		"Brasil": "🇧🇷", "Marruecos": "🇲🇦", "Haití": "🇭🇹", "Escocia": "🏴",
		"Estados Unidos": "🇺🇸", "Paraguay": "🇵🇾", "Australia": "🇦🇺", "Turquía": "🇹🇷",
		"Alemania": "🇩🇪", "Curazao": "🇨🇼", "Costa de Marfil": "🇨🇮", "Ecuador": "🇪🇨",
		"Países Bajos": "🇳🇱", "Japón": "🇯🇵", "Suecia": "🇸🇪", "Túnez": "🇹🇳",
		"Bélgica": "🇧🇪", "Egipto": "🇪🇬", "Irán": "🇮🇷", "Nueva Zelanda": "🇳🇿",
		"España": "🇪🇸", "Cabo Verde": "🇨🇻", "Arabia Saudita": "🇸🇦", "Uruguay": "🇺🇾",
		"Francia": "🇫🇷", "Senegal": "🇸🇳", "Irak": "🇮🇶", "Noruega": "🇳🇴",
		"Argentina": "🇦🇷", "Argelia": "🇩🇿", "Austria": "🇦🇹", "Jordania": "🇯🇴",
		"Portugal": "🇵🇹", "República Democrática del Congo": "🇨🇩", "Uzbekistán": "🇺🇿", "Colombia": "🇨🇴",
		"Inglaterra": "🏴", "Croacia": "🇭🇷", "Ghana": "🇬🇭", "Panamá": "🇵🇦",
	}

	matches := []struct {
		Home  string
		Away  string
		Group string
	}{
		{"México", "Sudáfrica", "A"}, {"México", "Corea del Sur", "A"}, {"México", "República Checa", "A"}, {"Sudáfrica", "Corea del Sur", "A"}, {"Sudáfrica", "República Checa", "A"}, {"Corea del Sur", "República Checa", "A"},
		{"Canadá", "Bosnia y Herzegovina", "B"}, {"Canadá", "Catar", "B"}, {"Canadá", "Suiza", "B"}, {"Bosnia y Herzegovina", "Catar", "B"}, {"Bosnia y Herzegovina", "Suiza", "B"}, {"Catar", "Suiza", "B"},
		{"Brasil", "Marruecos", "C"}, {"Brasil", "Haití", "C"}, {"Brasil", "Escocia", "C"}, {"Marruecos", "Haití", "C"}, {"Marruecos", "Escocia", "C"}, {"Haití", "Escocia", "C"},
		{"Estados Unidos", "Paraguay", "D"}, {"Estados Unidos", "Australia", "D"}, {"Estados Unidos", "Turquía", "D"}, {"Paraguay", "Australia", "D"}, {"Paraguay", "Turquía", "D"}, {"Australia", "Turquía", "D"},
		{"Alemania", "Curazao", "E"}, {"Alemania", "Costa de Marfil", "E"}, {"Alemania", "Ecuador", "E"}, {"Curazao", "Costa de Marfil", "E"}, {"Curazao", "Ecuador", "E"}, {"Costa de Marfil", "Ecuador", "E"},
		{"Países Bajos", "Japón", "F"}, {"Países Bajos", "Suecia", "F"}, {"Países Bajos", "Túnez", "F"}, {"Japón", "Suecia", "F"}, {"Japón", "Túnez", "F"}, {"Suecia", "Túnez", "F"},
		{"Bélgica", "Egipto", "G"}, {"Bélgica", "Irán", "G"}, {"Bélgica", "Nueva Zelanda", "G"}, {"Egipto", "Irán", "G"}, {"Egipto", "Nueva Zelanda", "G"}, {"Irán", "Nueva Zelanda", "G"},
		{"España", "Cabo Verde", "H"}, {"España", "Arabia Saudita", "H"}, {"España", "Uruguay", "H"}, {"Cabo Verde", "Arabia Saudita", "H"}, {"Cabo Verde", "Uruguay", "H"}, {"Arabia Saudita", "Uruguay", "H"},
		{"Francia", "Senegal", "I"}, {"Francia", "Irak", "I"}, {"Francia", "Noruega", "I"}, {"Senegal", "Irak", "I"}, {"Senegal", "Noruega", "I"}, {"Irak", "Noruega", "I"},
		{"Argentina", "Argelia", "J"}, {"Argentina", "Austria", "J"}, {"Argentina", "Jordania", "J"}, {"Argelia", "Austria", "J"}, {"Argelia", "Jordania", "J"}, {"Austria", "Jordania", "J"},
		{"Portugal", "República Democrática del Congo", "K"}, {"Portugal", "Uzbekistán", "K"}, {"Portugal", "Colombia", "K"}, {"República Democrática del Congo", "Uzbekistán", "K"}, {"República Democrática del Congo", "Colombia", "K"}, {"Uzbekistán", "Colombia", "K"},
		{"Inglaterra", "Croacia", "L"}, {"Inglaterra", "Ghana", "L"}, {"Inglaterra", "Panamá", "L"}, {"Croacia", "Ghana", "L"}, {"Croacia", "Panamá", "L"}, {"Ghana", "Panamá", "L"},
	}

	for _, m := range matches {
		_, err := r.DB.Exec("INSERT INTO matches (home_team, away_team, match_date, status, group_name, is_locked, home_flag, away_flag) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			m.Home, m.Away, "2026-06-15 00:00:00", "scheduled", m.Group, 0, flags[m.Home], flags[m.Away])
		if err != nil {
			return err
		}
	}
	log.Println("Full tournament seeded with flags successfully.")
	return nil
}

func (r *SQLiteRepository) GetAllMatches() ([]domain.Match, error) {
	rows, err := r.DB.Query("SELECT id, home_team, away_team, home_score, away_score, match_date, status, group_name, is_locked, home_flag, away_flag FROM matches")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []domain.Match
	for rows.Next() {
		var m domain.Match
		err := rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.HomeScore, &m.AwayScore, &m.Date, &m.Status, &m.Group, &m.IsLocked, &m.HomeFlag, &m.AwayFlag)
		if err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	return matches, nil
}

func (r *SQLiteRepository) GetMatchByID(id int) (*domain.Match, error) {
	var m domain.Match
	err := r.DB.QueryRow("SELECT id, home_team, away_team, home_score, away_score, match_date, status, group_name, is_locked, home_flag, away_flag FROM matches WHERE id = ?", id).
		Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.HomeScore, &m.AwayScore, &m.Date, &m.Status, &m.Group, &m.IsLocked, &m.HomeFlag, &m.AwayFlag)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *SQLiteRepository) UpdateMatchResult(id int, homeScore, awayScore int) error {
	_, err := r.DB.Exec("UPDATE matches SET home_score = ?, away_score = ?, status = 'finished' WHERE id = ?",
		homeScore, awayScore, id)
	return err
}

func (r *SQLiteRepository) ToggleLock(id int, locked bool) error {
	_, err := r.DB.Exec("UPDATE matches SET is_locked = ? WHERE id = ?", locked, id)
	return err
}

func (r *SQLiteRepository) LockAll(locked bool) error {
	val := 0
	if locked {
		val = 1
	}
	_, err := r.DB.Exec("UPDATE matches SET is_locked = ?", val)
	return err
}

func (r *SQLiteRepository) Save(p domain.Prediction) error {
	_, err := r.DB.Exec(`
		INSERT INTO predictions (user_id, match_id, home_score, away_score) 
		VALUES (?, ?, ?, ?) 
		ON CONFLICT(user_id, match_id) DO UPDATE SET home_score=excluded.home_score, away_score=excluded.away_score`,
		p.UserID, p.MatchID, p.HomeScore, p.AwayScore)
	return err
}

func (r *SQLiteRepository) GetByMatchID(matchID int) ([]domain.Prediction, error) {
	rows, err := r.DB.Query("SELECT id, user_id, match_id, home_score, away_score FROM predictions WHERE match_id = ?", matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var preds []domain.Prediction
	for rows.Next() {
		var p domain.Prediction
		err := rows.Scan(&p.ID, &p.UserID, &p.MatchID, &p.HomeScore, &p.AwayScore)
		if err != nil {
			return nil, err
		}
		preds = append(preds, p)
	}
	return preds, nil
}

func (r *SQLiteRepository) GetAllPredictions() ([]domain.Prediction, error) {
	rows, err := r.DB.Query("SELECT id, user_id, match_id, home_score, away_score FROM predictions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	preds := []domain.Prediction{}
	for rows.Next() {
		var p domain.Prediction
		err := rows.Scan(&p.ID, &p.UserID, &p.MatchID, &p.HomeScore, &p.AwayScore)
		if err != nil {
			return nil, err
		}
		preds = append(preds, p)
	}
	return preds, nil
}

func (r *SQLiteRepository) Create(u domain.User) error {
	_, err := r.DB.Exec("INSERT INTO users (username, email, password, is_admin) VALUES (?, ?, ?, ?)",
		u.Username, u.Email, u.Password, u.IsAdmin)
	return err
}

func (r *SQLiteRepository) GetByID(id int) (*domain.User, error) {
	var u domain.User
	err := r.DB.QueryRow("SELECT id, username, email, password, is_admin FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.IsAdmin)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *SQLiteRepository) GetByUsername(username string) (*domain.User, error) {
	var u domain.User
	err := r.DB.QueryRow("SELECT id, username, email, password, is_admin FROM users WHERE username = ?", username).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.IsAdmin)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *SQLiteRepository) GetByEmail(email string) (*domain.User, error) {
	var u domain.User
	err := r.DB.QueryRow("SELECT id, username, email, password, is_admin FROM users WHERE email = ?", email).
		Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.IsAdmin)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *SQLiteRepository) GetAll() ([]domain.User, error) {
	rows, err := r.DB.Query("SELECT id, username, email, password, is_admin FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.IsAdmin)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *SQLiteRepository) Delete(id int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM predictions WHERE user_id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM users WHERE id = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SQLiteRepository) Update(u domain.User) error {
	_, err := r.DB.Exec("UPDATE users SET username = ?, email = ?, is_admin = ? WHERE id = ?", u.Username, u.Email, u.IsAdmin, u.ID)
	return err
}
