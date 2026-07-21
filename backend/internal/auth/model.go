package auth

type CreateAuthRequest struct {
	UID   string `json:"uid" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type AuthLoginRequest struct {
	Email   string `json:"email" binding:"required,email"`
	IDToken string `json:"id_token" binding:"required"`
}

type AuthResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}
