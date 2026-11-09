package auth

type AuthLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type AuthResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
