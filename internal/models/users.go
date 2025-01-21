package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/mattn/go-sqlite3"
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
	salt := "random"
	hashedPassword, _ := hashPassword(password, salt)

	stmt := `INSERT INTO users (name, email, hashed_password, created) 
	         VALUES (?, ?, ?, DATETIME('now', 'localtime'))`

	_, err := m.DB.Exec(stmt, name, email, hashedPassword)
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
	var id int
	var hashedPassword string

	stmt := `SELECT id, hashed_password FROM users WHERE email = ?`

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	salt := "random"
	hashedPassword, _ = hashPassword(password, salt)
	hashedInputPassword, _ := hashPassword(password, salt)
	if hashedPassword != hashedInputPassword {
		return 0, ErrInvalidCredentials
	}
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
