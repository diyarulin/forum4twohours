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
	Likes     int
	Dislikes  int
	Author    string
	AuthorID  int
	Created   time.Time
	Status    string
}

// PostModel обёртка для соединения с базой данных
type PostModel struct {
	DB *sql.DB
}

// Insert добавляет новый пост в базу данных
func (m *PostModel) Insert(title, content, imagePath, category, author, status string, author_id int) (int, error) {
	// Категория и автор могут быть заданы по умолчанию
	// defaultCategory := "Uncategorized"
	// defaultAuthor := "Anonymous"
	stmt := `INSERT INTO posts (title, content, image_path, category, author, author_id, created, status) 
         VALUES (?, ?, ?, ?, ?, ?, DATETIME('now', 'localtime'), ?)` // 8 параметров!

	result, err := m.DB.Exec(
		stmt,
		title,     // 1
		content,   // 2
		imagePath, // 3
		category,  // 4
		author,    // 5
		author_id, // 6
		status,    // 8 (последний параметр)
	)

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
	stmt := `SELECT id, title, content, image_path, category, likes, dislikes, author, author_id, created, status FROM posts WHERE id = ?`

	row := m.DB.QueryRow(stmt, id)

	p := &Post{}
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.Category, &p.Likes, &p.Dislikes, &p.Author, &p.AuthorID, &p.Created, &p.Status)
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
	stmt := `SELECT id, title, content, image_path,  category, author, author_id, created FROM posts WHERE status = "approved" ORDER BY created DESC LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post

	for rows.Next() {
		p := &Post{}
		err = rows.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.Category, &p.Author, &p.AuthorID, &p.Created)
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
func (m *PostModel) UserPosts(userId int) ([]*Post, error) {
	stmt := `SELECT id, title, content, image_path,  category, author, author_id,  created FROM posts WHERE author_id = ?`

	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post

	for rows.Next() {
		p := &Post{}
		err = rows.Scan(&p.ID, &p.Title, &p.Content, &p.ImagePath, &p.Category, &p.Author, &p.AuthorID, &p.Created)
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
func (m *PostModel) UpdatePost(title, content, imagePath, category, author string, author_id, id int) error {
	// Категория и автор могут быть заданы по умолчанию
	// defaultCategory := "Uncategorized"
	// defaultAuthor := "Anonymous"
	stmt := `UPDATE posts SET title = ?, content = ?, image_path = ?, category = ?, author = ?, author_id = ? WHERE id = ?`

	_, err := m.DB.Exec(stmt, title, content, imagePath, category, author, author_id, id)
	if err != nil {
		return err
	}
	return nil
}
func (m *PostModel) DeletePost(id int) (string, error) {
	stmt1 := `SELECT image_path FROM posts WHERE id = ?`
	stmt2 := `DELETE FROM posts WHERE id = ?`
	var imagePath string
	err := m.DB.QueryRow(stmt1, id).Scan(&imagePath)
	if err != nil {
		return "", err
	}
	_, err = m.DB.Exec(stmt2, id)
	if err != nil {
		return "", err
	}
	return imagePath, nil
}

func (m *PostModel) SortByCategory(category string) ([]*Post, error) {
	stmt := `SELECT id, title, content, image_path, category, created, author, author_id FROM posts WHERE category = ? AND status = "approved"`
	rows, err := m.DB.Query(stmt, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []*Post
	for rows.Next() {
		post := &Post{}
		err = rows.Scan(&post.ID, &post.Title, &post.Content, &post.ImagePath, &post.Category, &post.Created, &post.Author, &post.AuthorID)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (m *PostModel) GetPendingPosts() ([]*Post, error) {
	stmt := `SELECT id, title, content, author, created 
             FROM posts 
             WHERE status = 'pending' 
             ORDER BY created DESC`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*Post
	for rows.Next() {
		p := &Post{}
		err = rows.Scan(&p.ID, &p.Title, &p.Content, &p.Author, &p.Created)
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

func (m *PostModel) ApprovePost(postID int) error {
	_, err := m.DB.Exec("UPDATE posts SET status = 'approved' WHERE id = ?", postID)
	return err
}
