package controller

import (
	"encoding/json"
	"net/http"

	"github.com/zhekagigs/golang_todo/logger"
	"github.com/zhekagigs/golang_todo/users"
)

type AuthHandler struct {
	UserStore *users.UserStore
}

func NewAuthHandler(userStore *users.UserStore) *AuthHandler {
	return &AuthHandler{UserStore: userStore}
}

type loginRequest struct {
	UserName string `json:"userName"`
}

func (ah *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var loginRequest *loginRequest

	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if handleError(w, err, http.StatusBadRequest, "error decoding request body") {
		return
	}
	var userId string
	var userName string
	userOld, exist := ah.UserStore.GetUser(loginRequest.UserName)

	if !exist {
		logger.Error.Println("User doesnt exist: ", loginRequest.UserName)
		userNew, err := ah.UserStore.AddUser(loginRequest.UserName)
		logger.Info.Println("User was added: ", userNew.UserName)
		if handleError(w, err, http.StatusBadRequest, "error add user") {
			return
		}
		userId = userNew.UserId.String()
		userName = userNew.UserName
	} else {
		userId = userOld.UserId.String()
		userName = userOld.UserName
	}
	logger.Error.Println("Found user: ", loginRequest.UserName)

	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    userId,
		MaxAge:   600,
		SameSite: http.SameSiteStrictMode,

		// SameSite: http.SameSiteDefaultMode,
		// Secure:   true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "UserName",
		Value:    userName,
		MaxAge:   600,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear UserName cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "UserName",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   true, // Set to true if using HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	// Clear Authorization cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true, // Keep this true for security
		Secure:   true, // Set to true if using HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout successful"})
}
