package main

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type sessionManager struct {
	db *sql.DB
}

func newSessionManager(db *sql.DB) *sessionManager {
	return &sessionManager{db: db}
}

type sessionData struct {
	UserID    string
	LastLogin time.Time
}

// Middleware for managing sessions
func (sm *sessionManager) LoadAndSave(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var session sessionData

		// Extract session token from the "Authorization" header
		sessionToken := r.Header.Get("Authorization")
		if sessionToken != "" {
			// Load session from the database
			var data []byte
			var expiry time.Time
			err := sm.db.QueryRow("SELECT data, expiry FROM sessions WHERE token = ?", sessionToken).Scan(&data, &expiry)
			if err == nil && time.Now().Before(expiry) {
				gob.NewDecoder(bytes.NewReader(data)).Decode(&session)
				// Attach session data to the request context if needed
				// For example: r = r.WithContext(context.WithValue(r.Context(), sessionKey, session))
			}
		}

		// Proceed with the request
		next.ServeHTTP(w, r)

		// Save the session to the database after handling the request
		if sessionToken == "" {
			sessionToken = uuid.New().String()
			w.Header().Set("Authorization", sessionToken)
		}

		var buf bytes.Buffer
		gob.NewEncoder(&buf).Encode(session)
		_, err := sm.db.Exec("INSERT OR REPLACE INTO sessions (token, data, expiry) VALUES (?, ?, ?)",
			sessionToken, buf.Bytes(), time.Now().Add(12*time.Hour))
		if err != nil {
			log.Printf("Failed to save session: %v", err)
		}
	})
}
