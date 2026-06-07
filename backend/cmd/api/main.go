package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"quiniela-backend/internal/application/services"
	"quiniela-backend/internal/infrastructure/handlers"
	"quiniela-backend/internal/infrastructure/repository"
)

func main() {
	driverName, connStr := buildDBConfig()
	lockWindow := buildLockWindow()
	qualifierLockAt := buildQualifierLockAt()

	repo, err := repository.NewSQLiteRepository(driverName, connStr, lockWindow)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer repo.DB.Close()

	matchService := services.NewMatchService(repo)
	predictionService := services.NewPredictionService(repo, repo, lockWindow)
	rankingService := services.NewRankingService(repo)
	userService := services.NewUserService(repo, repo)
	standingsService := services.NewStandingsService(repo)
	qualifierService := services.NewQualifierPredictionService(repo, repo, lockWindow, qualifierLockAt)

	matchHandler := &handlers.MatchHandler{Service: matchService}
	predictionHandler := &handlers.PredictionHandler{Service: predictionService}
	rankingHandler := &handlers.RankingHandler{
		Service:          rankingService,
		UserService:      userService,
		MatchRepo:        repo,
		PredRepo:         repo,
		QualifierService: qualifierService,
	}
	userHandler := &handlers.UserHandler{Service: userService}
	adminHandler := &handlers.AdminHandler{MatchService: matchService}
	standingsHandler := &handlers.StandingsHandler{Service: standingsService}
	qualifierHandler := &handlers.QualifierPredictionHandler{
		Service:     qualifierService,
		UserService: userService,
		MatchRepo:   repo,
	}
	configHandler := &handlers.ConfigHandler{
		QualifierLockAt:  qualifierLockAt,
		MatchLockWindow:  lockWindow,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /config", configHandler.Get)
	mux.HandleFunc("OPTIONS /config", handlers.HandleOptions)

	mux.HandleFunc("GET /matches", matchHandler.GetMatches)
	mux.HandleFunc("OPTIONS /matches", handlers.HandleOptions)

	mux.HandleFunc("POST /predictions", predictionHandler.CreatePrediction)
	mux.HandleFunc("OPTIONS /predictions", handlers.HandleOptions)
	mux.HandleFunc("GET /predictions", predictionHandler.GetPredictions)

	mux.HandleFunc("GET /ranking", rankingHandler.GetRanking)
	mux.HandleFunc("OPTIONS /ranking", handlers.HandleOptions)

	mux.HandleFunc("POST /register", userHandler.Register)
	mux.HandleFunc("OPTIONS /register", handlers.HandleOptions)
	mux.HandleFunc("POST /login", userHandler.Login)
	mux.HandleFunc("OPTIONS /login", handlers.HandleOptions)

	mux.HandleFunc("POST /admin/matches/result", adminHandler.UpdateMatchResult)
	mux.HandleFunc("OPTIONS /admin/matches/result", handlers.HandleOptions)
	mux.HandleFunc("POST /admin/matches/lock", adminHandler.ToggleLock)
	mux.HandleFunc("OPTIONS /admin/matches/lock", handlers.HandleOptions)
	mux.HandleFunc("POST /admin/matches/lock-all", adminHandler.ToggleLockAll)
	mux.HandleFunc("OPTIONS /admin/matches/lock-all", handlers.HandleOptions)
	mux.HandleFunc("DELETE /admin/users/{id}", userHandler.DeleteUser)
	mux.HandleFunc("OPTIONS /admin/users/{id}", handlers.HandleOptions)
	mux.HandleFunc("POST /admin/users/update", userHandler.UpdateUser)
	mux.HandleFunc("OPTIONS /admin/users/update", handlers.HandleOptions)

	mux.HandleFunc("GET /users/{id}/predictions", userHandler.GetUserPredictions)
	mux.HandleFunc("OPTIONS /users/{id}/predictions", handlers.HandleOptions)

	mux.HandleFunc("GET /standings", standingsHandler.GetAll)
	mux.HandleFunc("OPTIONS /standings", handlers.HandleOptions)

	mux.HandleFunc("GET /users/{id}/qualifier-predictions", qualifierHandler.GetByUser)
	mux.HandleFunc("OPTIONS /users/{id}/qualifier-predictions", handlers.HandleOptions)
	mux.HandleFunc("GET /groups/{group}/qualifier-predictions/me", qualifierHandler.GetMy)
	mux.HandleFunc("OPTIONS /groups/{group}/qualifier-predictions/me", handlers.HandleOptions)
	mux.HandleFunc("PUT /groups/{group}/qualifier-predictions", qualifierHandler.Upsert)
	mux.HandleFunc("OPTIONS /groups/{group}/qualifier-predictions", handlers.HandleOptions)

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func buildDBConfig() (driverName, connStr string) {
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

func buildLockWindow() time.Duration {
	raw := os.Getenv("LOCK_WINDOW_HOURS")
	if raw == "" {
		return 3 * time.Hour
	}
	hours, err := strconv.ParseFloat(raw, 64)
	if err != nil || hours <= 0 {
		log.Printf("warn: invalid LOCK_WINDOW_HOURS=%q, using default 3h", raw)
		return 3 * time.Hour
	}
	window := time.Duration(hours * float64(time.Hour))
	log.Printf("lock window configured: %s (%.4f hours)", window, hours)
	return window
}

func buildQualifierLockAt() time.Time {
	const defaultDeadline = "2026-06-11T18:00:00Z"
	raw := os.Getenv("QUALIFIER_LOCK_AT")
	if raw == "" {
		t, err := time.Parse(time.RFC3339, defaultDeadline)
		if err != nil {
			log.Fatalf("invalid default QUALIFIER_LOCK_AT=%q: %v", defaultDeadline, err)
		}
		log.Printf("qualifier lock deadline (default): %s", t.UTC().Format(time.RFC3339))
		return t.UTC()
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		log.Printf("warn: invalid QUALIFIER_LOCK_AT=%q, using default %s", raw, defaultDeadline)
		t, _ = time.Parse(time.RFC3339, defaultDeadline)
		return t.UTC()
	}
	log.Printf("qualifier lock deadline configured: %s", t.UTC().Format(time.RFC3339))
	return t.UTC()
}
