package main

import (
	"errors"
	"github.com/google/uuid"
	"net/http"
	"time"
)

func (app *application) setSession(w http.ResponseWriter, userID int) string {
	sessionID := uuid.New().String()
	app.sessions[sessionID] = userID

	http.SetCookie(w, &http.Cookie{
		Name:    "session_id",
		Value:   sessionID,
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour),
	})
	return sessionID
}

func (app *application) getCurrentUser(r *http.Request) (int, error) {
	// Получаем cookie сессии
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return 0, errors.New("no session found")
	}

	// Проверяем, существует ли сессия
	userID, exists := app.sessions[cookie.Value]
	if !exists {
		return 0, errors.New("invalid session")
	}

	return userID, nil
}

func (app *application) renewSessionToken(w http.ResponseWriter, r *http.Request) error {
	// Получаем текущую сессию
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return errors.New("no session found")
	}

	// Проверяем, существует ли текущая сессия
	userID, exists := app.sessions[cookie.Value]
	if !exists {
		return errors.New("invalid session")
	}

	// Удаляем старую сессию
	delete(app.sessions, cookie.Value)

	// Создаем новую сессию с новым токеном
	app.setSession(w, userID)

	return nil
}

func (app *application) deleteSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "No session found", http.StatusUnauthorized)
		return
	}

	// Проверка валидности сессии
	if _, exists := app.sessions[cookie.Value]; !exists {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Удаление сессии из хранилища
	delete(app.sessions, cookie.Value)

	// Удаление cookie на клиенте
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusOK)
	app.flash(w, r, "Logout successfully")
}
