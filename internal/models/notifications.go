package models

import (
	"database/sql"
	"time"
)

type Notification struct {
	ID        int
	UserID    int
	Type      string
	PostID    int
	CommentID int
	Created   time.Time
	IsRead    bool
	ActorID   int
	ActorName string
}

type NotificationModel struct {
	DB *sql.DB
}

func (m *NotificationModel) Insert(userID, actorID int, ntype string, postID, commentID int) error {
	stmt := `INSERT INTO notifications (user_id, type, post_id, comment_id, actor_id) VALUES (?, ?, ?, ?, ?)`

	_, err := m.DB.Exec(stmt, userID, ntype, postID, commentID, actorID)
	return err
}

func (m *NotificationModel) GetUnreadCount(userID int) (int, error) {
	var count int
	stmt := `SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0`
	err := m.DB.QueryRow(stmt, userID).Scan(&count)
	return count, err
}

func (m *NotificationModel) GetAll(userID int) ([]*Notification, error) {
	stmt := `SELECT n.id, n.type, n.post_id, n.comment_id, n.created, n.is_read,
         u.id, u.name
         FROM notifications n
         JOIN users u ON n.actor_id = u.id
         WHERE n.user_id = ?
         ORDER BY n.created DESC`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		var n Notification
		var isRead int
		err := rows.Scan(
			&n.ID, &n.Type, &n.PostID, &n.CommentID, &n.Created, &isRead,
			&n.ActorID, &n.ActorName,
		)
		if err != nil {
			return nil, err
		}
		n.IsRead = isRead == 1
		notifications = append(notifications, &n)
	}
	return notifications, nil
}

func (m *NotificationModel) MarkAllAsRead(userID int) error {
	stmt := `UPDATE notifications SET is_read = 1 
	WHERE user_id = ?`
	_, err := m.DB.Exec(stmt, userID)
	return err
}
