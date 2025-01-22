package models

import (
	"database/sql"
	"errors"
	"time"
)

// Post структура для хранения данных поста
type Post struct {
	ID        int
	Title     string
	Content   string
	ImagePath string
	Category  string
	Author    string
	Created   time.Time
}

// PostModel обёртка для соединения с базой данных
type PostModel struct {
	DB *sql.DB
}

// Insert добавляет новый пост в базу данных
func (m *PostModel) Insert(title, content, imagePath string) (int, error) {
	// Категория и автор могут быть заданы по умолчанию
	defaultCategory := "Uncategorized"
	defaultAuthor := "Anonymous"

	stmt := `INSERT INTO posts (title, content, image_path, category, author, created) 
	         VALUES (?, ?, ?, ?, ?, DATETIME('now', 'localtime'))`

	result, err := m.DB.Exec(stmt, title, content, imagePath, defaultCategory, defaultAuthor)
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
	stmt := `SELECT id, title, content, image_path, category, author, created FROM posts WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

	p := &Post{}
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.Category, &p.Author, &p.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	return p, nil
}

// Latest возвращает 10 последних постов
func (m *PostModel) Latest() ([]*Post, error) {
	stmt := `SELECT id, title, content, image_path,  category, author, created FROM posts ORDER BY created DESC LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post

	for rows.Next() {
		p := &Post{}
		err = rows.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.Category, &p.Author, &p.Created)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
func (m *PostModel) userPosts(user string) ([]*Post, error) {
	stmt := `SELECT id, title, content, image_path,  category,  created FROM posts WHERE author = ?`

	rows, err := m.DB.Query(stmt, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post

	for rows.Next() {
		p := &Post{}
		err = rows.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.Category, &p.Created)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
