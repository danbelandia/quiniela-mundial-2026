package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"quiniela-backend/internal/application/services"
	"quiniela-backend/internal/core/domain"
	"quiniela-backend/internal/core/ports"
	"strconv"
	"strings"
	"time"
)

func SetCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func HandleOptions(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	w.WriteHeader(http.StatusNoContent)
}

type MatchHandler struct {
	Service *services.MatchService
}

func (h *MatchHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	matches, err := h.Service.ListMatches()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(matches)
}

type PredictionHandler struct {
	Service *services.PredictionService
}

func (h *PredictionHandler) CreatePrediction(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p domain.Prediction
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.Service.PlacePrediction(p); err != nil {
		if errors.Is(err, services.ErrMatchLocked) {
			http.Error(w, "match is locked", http.StatusForbidden)
			return
		}
		log.Printf("Error saving prediction: %v", err)
		http.Error(w, "Failed to save", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *PredictionHandler) GetPredictions(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	preds, err := h.Service.GetAllPredictions()
	if err != nil {
		log.Printf("Error fetching predictions: %v", err)
		http.Error(w, "Failed to fetch predictions", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preds)
}

type RankingHandler struct {
	Service           *services.RankingService
	UserService       *services.UserService
	MatchRepo         ports.MatchRepository
	PredRepo          ports.PredictionRepository
	QualifierService  *services.QualifierPredictionService
	TopScorerService  *services.TopScorerService
}

func (h *RankingHandler) GetRanking(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	users, err := h.UserService.GetAllUsers()
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	if users == nil {
		users = []domain.User{}
	}

	matches, _ := h.MatchRepo.GetAllMatches()
	preds, _ := h.PredRepo.GetAllPredictions()

	matchMap := make(map[int]domain.Match)
	for _, m := range matches {
		matchMap[m.ID] = m
	}

	for i, u := range users {
		score := 0
		exactScore := 0
		winnerScore := 0
		for _, p := range preds {
			if p.UserID == u.ID {
				match, ok := matchMap[p.MatchID]
				if ok && match.Status == "finished" {
					pts := h.Service.CalculatePoints(p, match.HomeScore, match.AwayScore)
					score += pts
					if pts == 3 {
						exactScore++
					} else if pts == 2 {
						winnerScore++
					}
				}
			}
		}
		users[i].Score = score
		users[i].ExactScore = exactScore
		users[i].WinnerScore = winnerScore

		if h.QualifierService != nil {
			qScore, qErr := h.QualifierService.ScoreUser(u.ID)
			if qErr == nil {
				users[i].QualifierScore = qScore
			}
		}

		topScorerScore := 0
		if h.TopScorerService != nil {
			tsScore, tsErr := h.TopScorerService.ScoreUser(u.ID)
			if tsErr == nil {
				topScorerScore = tsScore
			}
		}
		users[i].TopScorerScore = topScorerScore
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

type UserHandler struct {
	Service *services.UserService
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	var creds struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	hasUser := creds.Username != ""
	hasEmail := creds.Email != ""
	if hasUser == hasEmail {
		http.Error(w, "Provide exactly one of: username or email", http.StatusBadRequest)
		return
	}

	identifier := creds.Username
	if hasEmail {
		identifier = creds.Email
	}

	user, err := h.Service.Login(identifier, creds.Password)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	var u domain.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.Service.Register(u); err != nil {
		log.Printf("Registration error: %v", err)
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
			http.Error(w, "El nombre de usuario ya está en uso", http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			http.Error(w, "El email ya está registrado", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to register: "+err.Error(), http.StatusInternalServerError)
		return
	}

	created, err := h.Service.Login(u.Username, u.Password)
	if err != nil {
		w.WriteHeader(http.StatusCreated)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *UserHandler) GetUserPredictions(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.Service.GetByID(id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	preds, err := h.Service.GetUserPredictions(id)
	if err != nil {
		log.Printf("Error fetching user predictions: %v", err)
		http.Error(w, "Failed to fetch predictions", http.StatusInternalServerError)
		return
	}
	if preds == nil {
		preds = []ports.PredictionWithMatch{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"user":        user,
		"predictions": preds,
	})
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	idStr := r.PathValue("id")
	id, _ := strconv.Atoi(idStr)
	if err := h.Service.DeleteUser(id); err != nil {
		log.Printf("Delete user error: %v", err)
		http.Error(w, "Failed to delete user: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	var u domain.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.Service.UpdateUser(u); err != nil {
		log.Printf("Update error: %v", err)
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type AdminHandler struct {
	MatchService *services.MatchService
}

func (h *AdminHandler) UpdateMatchResult(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	var data struct {
		ID        int `json:"id"`
		HomeScore int `json:"home_score"`
		AwayScore int `json:"away_score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.MatchService.UpdateResult(data.ID, data.HomeScore, data.AwayScore); err != nil {
		log.Printf("Error updating match result: %v", err)
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) ToggleLock(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	var data struct {
		ID       int  `json:"id"`
		IsLocked bool `json:"is_locked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.MatchService.ToggleLock(data.ID, data.IsLocked); err != nil {
		log.Printf("Error toggling lock: %v", err)
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) ToggleLockAll(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	var data struct {
		IsLocked bool `json:"is_locked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.MatchService.LockAll(data.IsLocked); err != nil {
		log.Printf("Error locking all: %v", err)
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type StandingsHandler struct {
	Service *services.StandingsService
}

func (h *StandingsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	standings, err := h.Service.CalculateAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(standings)
}

type QualifierPredictionHandler struct {
	Service     *services.QualifierPredictionService
	UserService *services.UserService
	MatchRepo   ports.MatchRepository
}

func (h *QualifierPredictionHandler) GetByUser(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}
	if _, err := h.UserService.GetByID(id); err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	views, err := h.Service.GetViewsForUser(id)
	if err != nil {
		log.Printf("Error fetching qualifier views for user %d: %v", id, err)
		http.Error(w, "Failed to fetch qualifier predictions", http.StatusInternalServerError)
		return
	}
	if views == nil {
		views = []ports.QualifierPredictionView{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(views)
}

func (h *QualifierPredictionHandler) GetMy(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	groupName := r.PathValue("group")
	userIDStr := r.URL.Query().Get("user_id")
	if groupName == "" {
		http.Error(w, "group is required", http.StatusBadRequest)
		return
	}
	if userIDStr == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	pred, err := h.Service.GetByUserAndGroup(userID, groupName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pred == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"group_name":       groupName,
			"predicted_first":  "",
			"predicted_second": "",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pred)
}

func (h *QualifierPredictionHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	groupName := r.PathValue("group")
	if groupName == "" {
		http.Error(w, "group is required", http.StatusBadRequest)
		return
	}

	var p domain.QualifierPrediction
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	p.GroupName = groupName

	if p.UserID == 0 || p.PredictedFirst == "" || p.PredictedSecond == "" {
		http.Error(w, "user_id, predicted_first and predicted_second are required", http.StatusBadRequest)
		return
	}
	if p.PredictedFirst == p.PredictedSecond {
		http.Error(w, "predicted_first and predicted_second must be different teams", http.StatusBadRequest)
		return
	}

	matches, err := h.MatchRepo.GetAllMatches()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	teamsInGroup := make(map[string]bool)
	for _, m := range matches {
		if m.Group == groupName {
			teamsInGroup[m.HomeTeam] = true
			teamsInGroup[m.AwayTeam] = true
		}
	}
	if !teamsInGroup[p.PredictedFirst] || !teamsInGroup[p.PredictedSecond] {
		http.Error(w, "predicted team is not in the group", http.StatusBadRequest)
		return
	}

	if err := h.Service.Upsert(p); err != nil {
		if errors.Is(err, services.ErrMatchLocked) {
			http.Error(w, "group is locked", http.StatusForbidden)
			return
		}
		log.Printf("Error saving qualifier prediction: %v", err)
		http.Error(w, "Failed to save: "+err.Error(), http.StatusInternalServerError)
		return
	}

	saved, err := h.Service.GetByUserAndGroup(p.UserID, groupName)
	if err != nil || saved == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(saved)
}

// ConfigHandler exposes global settings (lock deadlines, windows) so the
// frontend can align its UI state with the server.
type ConfigHandler struct {
	QualifierLockAt      time.Time
	MatchLockWindow      time.Duration
	TopScorerCandidates []domain.TopScorerCandidate
}

func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"qualifier_lock_at":       h.QualifierLockAt.UTC().Format(time.RFC3339),
		"match_lock_window_hours": h.MatchLockWindow.Hours(),
		"top_scorer_candidates":   h.TopScorerCandidates,
	})
}

// TopScorerHandler exposes the user's own top-scorer prediction.
type TopScorerHandler struct {
	Service *services.TopScorerService
}

func (h *TopScorerHandler) UpsertMine(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	var body struct {
		UserID          int    `json:"user_id"`
		PredictedPlayer string `json:"predicted_player"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if body.UserID == 0 || body.PredictedPlayer == "" {
		http.Error(w, "user_id and predicted_player are required", http.StatusBadRequest)
		return
	}
	if err := h.Service.Upsert(body.UserID, body.PredictedPlayer); err != nil {
		switch {
		case errors.Is(err, services.ErrMatchLocked):
			http.Error(w, "match is locked", http.StatusForbidden)
		case errors.Is(err, services.ErrInvalidPlayer):
			http.Error(w, "player not in candidate list", http.StatusBadRequest)
		default:
			log.Printf("Error saving top scorer prediction: %v", err)
			http.Error(w, "Failed to save", http.StatusInternalServerError)
		}
		return
	}
	saved, err := h.Service.GetMine(body.UserID)
	if err != nil || saved == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(saved)
}

func (h *TopScorerHandler) GetMine(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}
	pred, err := h.Service.GetMine(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pred == nil {
		http.Error(w, "no prediction", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pred)
}

func (h *TopScorerHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	pred, err := h.Service.GetByUserID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if pred == nil {
		http.Error(w, "no prediction", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pred)
}

// TopScorerAdminHandler exposes the admin endpoint to set the actual top scorer.
type TopScorerAdminHandler struct {
	Service *services.TopScorerService
}

func (h *TopScorerAdminHandler) SetActual(w http.ResponseWriter, r *http.Request) {
	SetCORS(w)
	var body struct {
		Player string `json:"player"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Player == "" {
		http.Error(w, "player is required", http.StatusBadRequest)
		return
	}
	if err := h.Service.SetActual(body.Player); err != nil {
		log.Printf("Error setting top scorer: %v", err)
		http.Error(w, "Failed to save", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
