package repository

import (
	"database/sql"
	"log"
	"time"
	"quiniela-backend/internal/core/domain"
	"quiniela-backend/internal/core/ports"

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
	CREATE TABLE IF NOT EXISTS group_qualifier_predictions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		group_name TEXT NOT NULL,
		predicted_first TEXT NOT NULL,
		predicted_second TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, group_name),
		FOREIGN KEY(user_id) REFERENCES users(id)
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
		"Brasil": "🇧🇷", "Marruecos": "🇲🇦", "Haití": "🇭🇹", "Escocia": "🏴󠁧󠁢󠁳󠁣󠁴󠁿",
		"Estados Unidos": "🇺🇸", "Paraguay": "🇵🇾", "Australia": "🇦🇺", "Turquía": "🇹🇷",
		"Alemania": "🇩🇪", "Curazao": "🇨🇼", "Costa de Marfil": "🇨🇮", "Ecuador": "🇪🇨",
		"Países Bajos": "🇳🇱", "Japón": "🇯🇵", "Suecia": "🇸🇪", "Túnez": "🇹🇳",
		"Bélgica": "🇧🇪", "Egipto": "🇪🇬", "Irán": "🇮🇷", "Nueva Zelanda": "🇳🇿",
		"España": "🇪🇸", "Cabo Verde": "🇨🇻", "Arabia Saudita": "🇸🇦", "Uruguay": "🇺🇾",
		"Francia": "🇫🇷", "Senegal": "🇸🇳", "Irak": "🇮🇶", "Noruega": "🇳🇴",
		"Argentina": "🇦🇷", "Argelia": "🇩🇿", "Austria": "🇦🇹", "Jordania": "🇯🇴",
		"Portugal": "🇵🇹", "República Democrática del Congo": "🇨🇩", "Uzbekistán": "🇺🇿", "Colombia": "🇨🇴",
		"Inglaterra": "🏴󠁧󠁢󠁥󠁮󠁧󠁿", "Croacia": "🇭🇷", "Ghana": "🇬🇭", "Panamá": "🇵🇦",
	}

	matches := []struct {
		Home  string
		Away  string
		Group string
		Date  string
	}{
		{"México", "Sudáfrica", "A", "2026-06-11T19:00:00Z"}, {"México", "Corea del Sur", "A", "2026-06-12T02:00:00Z"}, {"México", "República Checa", "A", "2026-06-12T19:00:00Z"}, {"Sudáfrica", "Corea del Sur", "A", "2026-06-13T03:00:00Z"}, {"Sudáfrica", "República Checa", "A", "2026-06-13T19:00:00Z"}, {"Corea del Sur", "República Checa", "A", "2026-06-13T21:00:00Z"},
		{"Canadá", "Bosnia y Herzegovina", "B", "2026-06-14T00:00:00Z"}, {"Canadá", "Catar", "B", "2026-06-14T03:00:00Z"}, {"Canadá", "Suiza", "B", "2026-06-14T17:00:00Z"}, {"Bosnia y Herzegovina", "Catar", "B", "2026-06-14T20:00:00Z"}, {"Bosnia y Herzegovina", "Suiza", "B", "2026-06-14T23:00:00Z"}, {"Catar", "Suiza", "B", "2026-06-15T02:00:00Z"},
		{"Brasil", "Marruecos", "C", "2026-06-15T16:00:00Z"}, {"Brasil", "Haití", "C", "2026-06-15T19:00:00Z"}, {"Brasil", "Escocia", "C", "2026-06-15T22:00:00Z"}, {"Marruecos", "Haití", "C", "2026-06-16T01:00:00Z"}, {"Marruecos", "Escocia", "C", "2026-06-16T19:00:00Z"}, {"Haití", "Escocia", "C", "2026-06-16T22:00:00Z"},
		{"Estados Unidos", "Paraguay", "D", "2026-06-17T01:00:00Z"}, {"Estados Unidos", "Australia", "D", "2026-06-17T04:00:00Z"}, {"Estados Unidos", "Turquía", "D", "2026-06-17T17:00:00Z"}, {"Paraguay", "Australia", "D", "2026-06-17T20:00:00Z"}, {"Paraguay", "Turquía", "D", "2026-06-17T23:00:00Z"}, {"Australia", "Turquía", "D", "2026-06-18T02:00:00Z"},
		{"Alemania", "Curazao", "E", "2026-06-18T16:00:00Z"}, {"Alemania", "Costa de Marfil", "E", "2026-06-18T19:00:00Z"}, {"Alemania", "Ecuador", "E", "2026-06-18T22:00:00Z"}, {"Curazao", "Costa de Marfil", "E", "2026-06-19T01:00:00Z"}, {"Curazao", "Ecuador", "E", "2026-06-19T19:00:00Z"}, {"Costa de Marfil", "Ecuador", "E", "2026-06-19T22:00:00Z"},
		{"Países Bajos", "Japón", "F", "2026-06-20T01:00:00Z"}, {"Países Bajos", "Suecia", "F", "2026-06-20T04:00:00Z"}, {"Países Bajos", "Túnez", "F", "2026-06-20T17:00:00Z"}, {"Japón", "Suecia", "F", "2026-06-20T20:00:00Z"}, {"Japón", "Túnez", "F", "2026-06-21T02:00:00Z"}, {"Suecia", "Túnez", "F", "2026-06-21T04:00:00Z"},
		{"Bélgica", "Egipto", "G", "2026-06-21T16:00:00Z"}, {"Bélgica", "Irán", "G", "2026-06-21T19:00:00Z"}, {"Bélgica", "Nueva Zelanda", "G", "2026-06-21T22:00:00Z"}, {"Egipto", "Irán", "G", "2026-06-22T01:00:00Z"}, {"Egipto", "Nueva Zelanda", "G", "2026-06-22T17:00:00Z"}, {"Irán", "Nueva Zelanda", "G", "2026-06-22T21:00:00Z"},
		{"España", "Cabo Verde", "H", "2026-06-23T00:00:00Z"}, {"España", "Arabia Saudita", "H", "2026-06-23T03:00:00Z"}, {"España", "Uruguay", "H", "2026-06-23T17:00:00Z"}, {"Cabo Verde", "Arabia Saudita", "H", "2026-06-23T20:00:00Z"}, {"Cabo Verde", "Uruguay", "H", "2026-06-23T23:00:00Z"}, {"Arabia Saudita", "Uruguay", "H", "2026-06-24T02:00:00Z"},
		{"Francia", "Senegal", "I", "2026-06-24T19:00:00Z"}, {"Francia", "Irak", "I", "2026-06-24T19:00:00Z"}, {"Francia", "Noruega", "I", "2026-06-24T22:00:00Z"}, {"Senegal", "Irak", "I", "2026-06-24T22:00:00Z"}, {"Senegal", "Noruega", "I", "2026-06-25T01:00:00Z"}, {"Irak", "Noruega", "I", "2026-06-25T01:00:00Z"},
		{"Argentina", "Argelia", "J", "2026-06-25T20:00:00Z"}, {"Argentina", "Austria", "J", "2026-06-25T20:00:00Z"}, {"Argentina", "Jordania", "J", "2026-06-25T23:00:00Z"}, {"Argelia", "Austria", "J", "2026-06-25T23:00:00Z"}, {"Argelia", "Jordania", "J", "2026-06-26T02:00:00Z"}, {"Austria", "Jordania", "J", "2026-06-26T02:00:00Z"},
		{"Portugal", "República Democrática del Congo", "K", "2026-06-26T19:00:00Z"}, {"Portugal", "Uzbekistán", "K", "2026-06-26T19:00:00Z"}, {"Portugal", "Colombia", "K", "2026-06-27T00:00:00Z"}, {"República Democrática del Congo", "Uzbekistán", "K", "2026-06-27T00:00:00Z"}, {"República Democrática del Congo", "Colombia", "K", "2026-06-27T03:00:00Z"}, {"Uzbekistán", "Colombia", "K", "2026-06-27T03:00:00Z"},
		{"Inglaterra", "Croacia", "L", "2026-06-27T21:00:00Z"}, {"Inglaterra", "Ghana", "L", "2026-06-27T21:00:00Z"}, {"Inglaterra", "Panamá", "L", "2026-06-27T23:30:00Z"}, {"Croacia", "Ghana", "L", "2026-06-27T23:30:00Z"}, {"Croacia", "Panamá", "L", "2026-06-28T02:00:00Z"}, {"Ghana", "Panamá", "L", "2026-06-28T02:00:00Z"},
	}

	for _, m := range matches {
		_, err := r.DB.Exec("INSERT INTO matches (home_team, away_team, match_date, status, group_name, is_locked, home_flag, away_flag) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			m.Home, m.Away, m.Date, "scheduled", m.Group, 0, flags[m.Home], flags[m.Away])
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
		var matchDateStr string
		err := rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.HomeScore, &m.AwayScore, &matchDateStr, &m.Status, &m.Group, &m.IsLocked, &m.HomeFlag, &m.AwayFlag)
		if err != nil {
			return nil, err
		}
		m.Date = parseMatchDate(matchDateStr)
		matches = append(matches, m)
	}
	return matches, nil
}

func (r *SQLiteRepository) GetMatchByID(id int) (*domain.Match, error) {
	var m domain.Match
	var matchDateStr string
	err := r.DB.QueryRow("SELECT id, home_team, away_team, home_score, away_score, match_date, status, group_name, is_locked, home_flag, away_flag FROM matches WHERE id = ?", id).
		Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.HomeScore, &m.AwayScore, &matchDateStr, &m.Status, &m.Group, &m.IsLocked, &m.HomeFlag, &m.AwayFlag)
	if err != nil {
		return nil, err
	}
	m.Date = parseMatchDate(matchDateStr)
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

func (r *SQLiteRepository) GetByUserIDWithMatch(userID int) ([]ports.PredictionWithMatch, error) {
	rows, err := r.DB.Query(`
		SELECT
			m.id, m.group_name, m.home_team, m.home_flag, m.away_team, m.away_flag, m.match_date, m.status, m.home_score, m.away_score,
			p.id, p.user_id, p.match_id, p.home_score, p.away_score
		FROM matches m
		LEFT JOIN predictions p ON p.match_id = m.id AND p.user_id = ?
		ORDER BY m.group_name, m.match_date`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ports.PredictionWithMatch
	for rows.Next() {
		var (
			matchID, realHome, realAway       int
			group, homeTeam, homeFlag         string
			awayTeam, awayFlag, matchDate, status string
			predID, predUserID, predMatchID   sql.NullInt64
			predHome, predAway                sql.NullInt64
		)
		if err := rows.Scan(
			&matchID, &group, &homeTeam, &homeFlag, &awayTeam, &awayFlag, &matchDate, &status, &realHome, &realAway,
			&predID, &predUserID, &predMatchID, &predHome, &predAway,
		); err != nil {
			return nil, err
		}
		pwm := ports.PredictionWithMatch{
			GroupName: group,
			HomeTeam:  homeTeam,
			HomeFlag:  homeFlag,
			AwayTeam:  awayTeam,
			AwayFlag:  awayFlag,
			MatchDate: matchDate,
			Status:    status,
			RealHome:  realHome,
			RealAway:  realAway,
		}
		if predID.Valid {
			pwm.HasPrediction = true
			pwm.PredictionID = int(predID.Int64)
			pwm.PredHome = int(predHome.Int64)
			pwm.PredAway = int(predAway.Int64)
			p := domain.Prediction{
				HomeScore: pwm.PredHome,
				AwayScore: pwm.PredAway,
			}
			if status == "finished" {
				pwm.Points = p.PointsEarned(pwm.RealHome, pwm.RealAway)
			}
		}
		out = append(out, pwm)
	}
	return out, nil
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

func parseMatchDate(s string) time.Time {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05.999999-07:00",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05.999999+00:00",
		"2006-01-02 15:04:05+00:00",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	log.Printf("warn: could not parse match_date: %q", s)
	return time.Time{}
}

func (r *SQLiteRepository) UpsertQualifierPrediction(p domain.QualifierPrediction) error {
	_, err := r.DB.Exec(`
		INSERT INTO group_qualifier_predictions (user_id, group_name, predicted_first, predicted_second)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id, group_name) DO UPDATE SET
			predicted_first = excluded.predicted_first,
			predicted_second = excluded.predicted_second,
			created_at = CURRENT_TIMESTAMP`,
		p.UserID, p.GroupName, p.PredictedFirst, p.PredictedSecond)
	return err
}

func (r *SQLiteRepository) GetQualifierPredictionByUserAndGroup(userID int, groupName string) (*domain.QualifierPrediction, error) {
	row := r.DB.QueryRow(`
		SELECT id, user_id, group_name, predicted_first, predicted_second, created_at
		FROM group_qualifier_predictions
		WHERE user_id = ? AND group_name = ?`, userID, groupName)
	var p domain.QualifierPrediction
	var createdAtStr string
	err := row.Scan(&p.ID, &p.UserID, &p.GroupName, &p.PredictedFirst, &p.PredictedSecond, &createdAtStr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.CreatedAt = parseMatchDate(createdAtStr)
	return &p, nil
}

func (r *SQLiteRepository) GetAllQualifierPredictionsByUser(userID int) ([]domain.QualifierPrediction, error) {
	rows, err := r.DB.Query(`
		SELECT id, user_id, group_name, predicted_first, predicted_second, created_at
		FROM group_qualifier_predictions WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var preds []domain.QualifierPrediction
	for rows.Next() {
		var p domain.QualifierPrediction
		var createdAtStr string
		if err := rows.Scan(&p.ID, &p.UserID, &p.GroupName, &p.PredictedFirst, &p.PredictedSecond, &createdAtStr); err != nil {
			return nil, err
		}
		p.CreatedAt = parseMatchDate(createdAtStr)
		preds = append(preds, p)
	}
	return preds, nil
}

func (r *SQLiteRepository) GetAllQualifierPredictions() ([]domain.QualifierPrediction, error) {
	rows, err := r.DB.Query(`
		SELECT id, user_id, group_name, predicted_first, predicted_second, created_at
		FROM group_qualifier_predictions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var preds []domain.QualifierPrediction
	for rows.Next() {
		var p domain.QualifierPrediction
		var createdAtStr string
		if err := rows.Scan(&p.ID, &p.UserID, &p.GroupName, &p.PredictedFirst, &p.PredictedSecond, &createdAtStr); err != nil {
			return nil, err
		}
		p.CreatedAt = parseMatchDate(createdAtStr)
		preds = append(preds, p)
	}
	return preds, nil
}
