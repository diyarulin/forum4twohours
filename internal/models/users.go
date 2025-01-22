package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword string
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

func (m *UserModel) getAuthor(id string) (string, error) {
	stmt := "SELECT name FROM users WHERE id = ?"
	row := m.DB.QueryRow(stmt, id)
	var name string
	if err := row.Scan(&name); err != nil {
		return "", err
	}

	return name, nil
}
