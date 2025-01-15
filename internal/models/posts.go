package models

import (
	"database/sql"
	"errors"
	"time"
)

// Post структура для хранения данных поста
type Post struct {
	ID      int
	Title   string
	Content string
	Created time.Time
}

// PostModel обёртка для соединения с базой данных
type PostModel struct {
	DB *sql.DB
}

// Insert добавляет новый пост в базу данных
func (m *PostModel) Insert(title string, content string) (int, error) {
	stmt := `INSERT INTO posts (title, content, created)
             VALUES (?, ?, DATETIME('now'))`

	result, err := m.DB.Exec(stmt, title, content)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// Get возвращает пост по ID
func (m *PostModel) Get(id int) (*Post, error) {
	stmt := `SELECT id, title, content, created FROM posts WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

	p := &Post{}

	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("record not found")
		}
		return nil, err
	}

	return p, nil
}

// Latest возвращает 10 последних постов
func (m *PostModel) Latest() ([]*Post, error) {
	stmt := `SELECT id, title, content, created FROM posts ORDER BY id DESC LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []*Post{}

	for rows.Next() {
		s := &Post{}
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created)
		if err != nil {
			return nil, err
		}
		posts = append(posts, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
