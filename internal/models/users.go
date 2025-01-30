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
	Role           string
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users
    (name, email, hashed_password, created, role) 
	         VALUES (?, ?, ?, DATETIME('now', 'localtime'), 'user')`

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		var sqliteError *sqlite3.Error
		// Проверяем, если ошибка связана с дублированием email
		if errors.As(err, &sqliteError) {
			// SQLite может возвращать ошибки, связанные с нарушением уникальности
			// В этом примере мы ищем ошибку по тексту
			if strings.Contains(sqliteError.Error(), "UNIQUE constraint failed: users"+
				".email") {
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

	stmt := "SELECT id, hashed_password FROM users" +
		" WHERE email = ?"

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
	stmt := `SELECT id, name, email, hashed_password, created, role FROM users WHERE id = ?`
	row := m.DB.QueryRow(stmt, id)

	u := &User{}
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.HashedPassword, &u.Created, &u.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	return u, nil
}
func (m *UserModel) GetAllUsers() ([]*User, error) {
	stmt := `SELECT id, name, email, role FROM users
`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		err = rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m *UserModel) UpdatePassword(hashedPassword string, id int) error {
	stmt := "UPDATE users" +
		" SET hashed_password = ? WHERE id = ?"
	_, err := m.DB.Exec(stmt, hashedPassword, id)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) GetOrCreateOAuthUser(email, name, provider, provider_id string) (int, error) {
	var userID int

	// Проверяем, существует ли пользователь с данным oauth_id и провайдером
	err := m.DB.QueryRow(`
        SELECT id FROM users
                  
        WHERE provider = ? AND provider_id = ?
    `, provider, provider_id).Scan(&userID)

	if err == nil {
		return userID, nil // Пользователь существует
	}

	// Если пользователя нет, создаем нового
	result, err := m.DB.Exec(`
        INSERT INTO users
            (name, email, provider, provider_id, role, created, hashed_password) 
        VALUES (?, ?, ?, ?, 'user', DATETIME('now'), 'google')
    `, name, email, provider, provider_id)

	if err != nil {
		return 0, err
	}

	id, _ := result.LastInsertId()
	return int(id), nil
}

func (m *UserModel) PromoteUser(userID int) error {
	_, err := m.DB.Exec("UPDATE users"+
		" SET role = 'moderator' WHERE id = ?", userID)
	return err
}

func (m *UserModel) DemoteUser(userID int) error {
	_, err := m.DB.Exec("UPDATE users"+
		" SET role = 'user' WHERE id = ?", userID)
	return err
}

func (m *UserModel) ApplyForModerator(userID int) error {
	stmt := `UPDATE users SET role = "pending_moderator" WHERE id = ? AND role = "user"`
	_, err := m.DB.Exec(stmt, userID)
	return err
}

func (m *UserModel) GetPendingModerators() ([]*User, error) {
	stmt := `SELECT id, name, email, role FROM users WHERE role = 'pending_moderator' OR role = 'moderator'`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
