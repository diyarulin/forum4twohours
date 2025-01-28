package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword string
	Provider       string
	ProviderID     string
	Created        time.Time
}


type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created) 
	         VALUES (?, ?, ?, DATETIME('now', 'localtime'))`

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		var sqliteError *sqlite3.Error
		// Проверяем, если ошибка связана с дублированием email
		if errors.As(err, &sqliteError) {
			// SQLite может возвращать ошибки, связанные с нарушением уникальности
			// В этом примере мы ищем ошибку по тексту
			if strings.Contains(sqliteError.Error(), "UNIQUE constraint failed: users.email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	// Retrieve the id and hashed password associated with the given email. If
	// no matching email exists we return the ErrInvalidCredentials error.
	var id int
	var hashedPassword []byte

	stmt := "SELECT id, hashed_password FROM users WHERE email = ?"

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// Check whether the hashed password and plain-text password provided match.
	// If they don't, we return the ErrInvalidCredentials error.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// Otherwise, the password is correct. Return the user ID.
	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	return false, nil
}

func hashPassword(password string, salt string) (string, error) {
	h := sha256.New()
	h.Write([]byte(password + salt))
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (m *UserModel) Get(id int) (*User, error) {
	stmt := `SELECT name, email, hashed_password, created FROM users WHERE id = ?`
	row := m.DB.QueryRow(stmt, id)
	u := &User{}
	err := row.Scan(&u.Name, &u.Email, &u.HashedPassword, &u.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	return u, nil
}
func (m *UserModel) UpdatePassword(hashedPassword string, id int) error {
	stmt := "UPDATE users SET hashed_password = ? WHERE id = ?"
	_, err := m.DB.Exec(stmt, hashedPassword, id)
	if err != nil {
		return err
	}
	return nil
}


func (m *UserModel) GetOrCreateOAuthUser(email, name, provider, providerID string) (int, error) {
	var id int

	// Проверяем, существует ли пользователь с таким email
	query := `SELECT id FROM users WHERE email = ?`
	err := m.DB.QueryRow(query, email).Scan(&id)

	if err == sql.ErrNoRows {
		// Если пользователь с таким email не найден, создаём нового пользователя
		stmt := `INSERT INTO users (name, email, provider, provider_id, hashed_password, created) 
		         VALUES (?, ?, ?, ?, "", CURRENT_TIMESTAMP)`
		result, err := m.DB.Exec(stmt, name, email, provider, providerID)
		if err != nil {
			return 0, err
		}

		userID, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}

		id = int(userID)
	} else if err != nil {
		return 0, err
	} else {
		// Если пользователь с таким email уже существует, проверяем provider и provider_id
		// Если они уже связаны, возвращаем существующий id
		var existingProvider string
		var existingProviderID string
		query := `SELECT provider, provider_id FROM users WHERE id = ?`
		err := m.DB.QueryRow(query, id).Scan(&existingProvider, &existingProviderID)
		if err != nil {
			return 0, err
		}

		// Если аккаунт уже ассоциирован с данным провайдером, возвращаем его
		if existingProvider == provider && existingProviderID == providerID {
			return id, nil
		}

		// Если провайдер или ID провайдера изменился, обновляем их
		_, err = m.DB.Exec(`UPDATE users SET provider = ?, provider_id = ? WHERE id = ?`, provider, providerID, id)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}
