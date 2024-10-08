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

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request, refreshToken string, user *database.User) {
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

// 2. Validate the JWT token and load the associated user
func (cfg *apiConfig) requireValidJWTToken(next func(w http.ResponseWriter, r *http.Request, token string, user *database.User)) func(w http.ResponseWriter, r *http.Request, token string) {
	return func(w http.ResponseWriter, r *http.Request, token string) {
		userId, err := auth.ValidateJWT(token, cfg.jwtSecret) // Validate JWT token
		if err != nil {
			responseWithJsonError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Retrieve user from database using the userId extracted from the token
		user, err := cfg.db.GetUserById(r.Context(), userId)
		if err != nil {
			responseWithJsonError(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Call the next function with token and user
		next(w, r, token, &user)
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
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func mapToJson(du *database.User, token string, refreshToken string) UserResponseLogin {
	return UserResponseLogin{
		ID:           du.ID,
		Email:        du.Email,
		CreatedAt:    du.CreatedAt,
		UpdatedAt:    du.UpdatedAt,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  du.IsChirpyRed,
	}
}

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request, token string) {
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

	userId, err := auth.ValidateJWT(token, cfg.jwtSecret)
	user, err := cfg.db.GetUserById(r.Context(), userId)
	if err != nil {
		responseWithJsonError(w, err.Error(), 401)
		return
	}

	if userId != user.ID {
		responseWithJsonError(w, "Unauthorized", 401)
		return
	}
	const minEntropy = 1
	if err := passwordvalidator.Validate(userReq.Password, minEntropy); err != nil {
		responseWithJsonError(w, err.Error(), 400)
	}
	hashedPasswd, err := auth.HashPassword(userReq.Password)

	updatedUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          userReq.Email,
		HashedPassword: hashedPasswd,
		ID:             user.ID,
	})
	if err != nil {
		responseWithJsonError(w, err.Error(), 500)
		return
	}

	responseWithJson(mapToJson(&updatedUser, "", ""), w, 200)

}
func (cfg *apiConfig) handleUpgradePolkaUser(w http.ResponseWriter, r *http.Request) {
	type eventReqBody struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		responseWithJsonError(w, err.Error(), 401)
		return
	}
	if apiKey != getEnvVariable("POLKA_KEY") {
		responseWithJsonError(w, "Unauthorized", 401)
		return
	}

	var eventReq eventReqBody
	json.NewDecoder(r.Body).Decode(&eventReq)
	if eventReq.Event == "" {
		responseWithJsonError(w, "Event is required", 400)
		return
	}
	if eventReq.Data.UserId == "" {
		responseWithJsonError(w, "User ID is required", 400)
		return
	}
	if eventReq.Event != "user.upgraded" {
		responseWithJsonError(w, "we dont know how to handle other events", 204)
		return
	}
	userId, err := uuid.Parse(eventReq.Data.UserId)
	if err != nil {
		responseWithJsonError(w, "Invalid user ID", 404)
		return
	}
	_, err = cfg.db.UpdateIsChirpyRed(r.Context(), database.UpdateIsChirpyRedParams{
		IsChirpyRed: true,
		ID:          userId,
	})
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			responseWithJsonError(w, "User not found", 404)
			return
		}
		responseWithJsonError(w, err.Error(), 500)
		return
	}
	w.WriteHeader(204)
}
