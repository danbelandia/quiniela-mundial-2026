package domain

type User struct {
	ID            int    `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	Score         int    `json:"score"`
	QualifierScore int   `json:"qualifier_score"`
	IsAdmin       bool   `json:"is_admin"`
}
