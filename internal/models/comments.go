package models

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID      int
	PostID  int
	Content string
	Author  string
	Created time.Time
}
type CommentModel struct {
	DB *sql.DB
}

func (m *CommentModel) GetByPostID(postID int) ([]*Comment, error) {
	stmt := `SELECT id, post_id, content, author, created FROM comments WHERE post_id = ? ORDER BY created ASC`

	rows, err := m.DB.Query(stmt, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		comment := &Comment{}
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.Content, &comment.Author, &comment.Created)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (m *CommentModel) Insert(comment *Comment) error {
	stmt := `INSERT INTO comments (post_id, content, author, created) VALUES (?, ?, ?, DATETIME('now', 'localtime'))`

	// Используем Exec для вставки данных
	result, err := m.DB.Exec(stmt, comment.PostID, comment.Content, comment.Author)
	if err != nil {
		return err
	}

	// Получаем последний вставленный ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Устанавливаем ID у комментария
	comment.ID = int(id)

	return nil
}
func (m *CommentModel) Delete(commentID int) error {
	stmt := `DELETE FROM comments WHERE id = ?`

	_, err := m.DB.Exec(stmt, commentID)
	if err != nil {
		return err
	}

	return nil
}
func (m *CommentModel) Update(commentID int, content string) error {
	stmt := `UPDATE comments SET content = ?, created = UTC_TIMESTAMP() WHERE id = ?`

	_, err := m.DB.Exec(stmt, content, commentID)
	if err != nil {
		return err
	}

	return nil
}
func (m *CommentModel) GetByID(commentID int) (*Comment, error) {
	stmt := `SELECT id, post_id, content, author, created FROM comments WHERE id = ?`

	row := m.DB.QueryRow(stmt, commentID)

	comment := &Comment{}
	err := row.Scan(&comment.ID, &comment.PostID, &comment.Content, &comment.Author, &comment.Created)
	if err != nil {
		return nil, err
	}

	return comment, nil
}
