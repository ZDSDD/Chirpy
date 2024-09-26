package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ZDSDD/Chirpy/internal/auth"
	"github.com/ZDSDD/Chirpy/internal/database"
	"github.com/google/uuid"
	passwordvalidator "github.com/wagslane/go-password-validator"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type UserReqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var userReq UserReqBody
	json.NewDecoder(r.Body).Decode(&userReq)
	if userReq.Email == "" {
		responseWithJsonError(w, "Email is required", 400)
		return
	}
	if userReq.Password == "" {
		responseWithJsonError(w, "Password is required", 400)
		return
	}
	user, err := cfg.db.GetUserByEmail(r.Context(), userReq.Email)
	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}

	if err := auth.CheckPasswordHash(userReq.Password, user.HashedPassword); err != nil {
		responseWithJsonError(w, "Invalid password", 401)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}
	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})

	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}

	responseWithJson(mapToJson(&user, token, refreshToken), w, http.StatusOK)
}

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request, refreshToken string) {
	rtdb, err := cfg.db.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		responseWithJsonError(w, err.Error(), 401)
		return
	}
	if rtdb.ExpiresAt.Before(time.Now()) {
		responseWithJsonError(w, "Refresh token expired", 401)
		return
	}
	if rtdb.RevokedAt.Valid {
		responseWithJsonError(w, "Refresh token revoked", 401)
		return
	}
	user, err := cfg.db.GetUserById(r.Context(), rtdb.UserID)
	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}
	responseWithJson(map[string]string{"token": token}, w, http.StatusOK)
}
func (cfg *apiConfig) requireBearerToken(next func(w http.ResponseWriter, r *http.Request, token string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			responseWithJsonError(w, err.Error(), 401)
			return
		}
		if token == "" {
			responseWithJsonError(w, "bearer token is required", 400)
			return
		}
		next(w, r, token)
	}
}
func (cfg *apiConfig) handleRevokeToken(w http.ResponseWriter, r *http.Request, refreshToken string) {

	err := cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}
func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type UserReqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var userReq UserReqBody
	json.NewDecoder(r.Body).Decode(&userReq)
	if userReq.Email == "" {
		responseWithJsonError(w, "Email is required", 400)
		return
	}
	if userReq.Password == "" {
		responseWithJsonError(w, "Password is required", 400)
		return
	}
	const minEntropy = 1
	if err := passwordvalidator.Validate(userReq.Password, minEntropy); err != nil {
		responseWithJsonError(w, err.Error(), 400)
	}
	hashedPasswd, err := auth.HashPassword(userReq.Password)

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          userReq.Email,
		HashedPassword: hashedPasswd,
	})
	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}
	responseWithJson(mapToJson(&user, "", ""), w, http.StatusCreated)
}

type UserResponseLogin struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

func mapToJson(du *database.User, token string, refreshToken string) UserResponseLogin {
	return UserResponseLogin{
		ID:           du.ID,
		Email:        du.Email,
		CreatedAt:    du.CreatedAt,
		UpdatedAt:    du.UpdatedAt,
		Token:        token,
		RefreshToken: refreshToken,
	}
}
