package main

import (
	"errors"
	"github.com/google/uuid"
	"net/http"
	"time"
)

func (app *application) setSession(w http.ResponseWriter, userID int) string {

	app.mu.Lock()
	defer app.mu.Unlock()

	for sessionID, uid := range app.sessions {
		if uid == userID {
			delete(app.sessions, sessionID)
		}
	}

	sessionID := uuid.New().String()
	app.sessions[sessionID] = userID

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	return sessionID
}

func (app *application) getCurrentUser(r *http.Request) (int, error) {
	app.mu.Lock()
	defer app.mu.Unlock()

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
	app.mu.Lock()
	defer app.mu.Unlock()

	// Получаем текущую сессию
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return errors.New("no session found")
	}

	// Проверяем существование сессии
	userID, exists := app.sessions[cookie.Value]
	if !exists {
		return errors.New("invalid session")
	}

	// Удаляем ВСЕ сессии пользователя
	for sessionID, uid := range app.sessions {
		if uid == userID {
			delete(app.sessions, sessionID)
		}
	}

	// Создаем новую сессию
	newSessionID := uuid.New().String()
	app.sessions[newSessionID] = userID

	// Обновляем cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    newSessionID,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	return nil
}

func (app *application) deleteSession(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()
	defer app.mu.Unlock()

	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "No session found", http.StatusUnauthorized)
		return
	}

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

	app.flash(w, r, "Logout successfully")
}
