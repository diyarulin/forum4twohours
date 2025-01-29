package models

import (
	"database/sql"
)

type ReactionModel struct {
	DB *sql.DB
}

func (m *ReactionModel) isLiked(postID, userID int) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id = ? AND user_id = ?)`
	err := m.DB.QueryRow(stmt, postID, userID).Scan(&exists)
	return exists, err
}

func (m *ReactionModel) isDisliked(postID, userID int) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT 1 FROM post_dislikes WHERE post_id = ? AND user_id = ?)`
	err := m.DB.QueryRow(stmt, postID, userID).Scan(&exists)
	return exists, err
}

func (m *ReactionModel) LikePost(postID, userID int) error {
	liked, err := m.isLiked(postID, userID)
	if err != nil {
		return err
	}
	disliked, err := m.isDisliked(postID, userID)
	if err != nil {
		return err
	}

	if liked {
		return m.RemoveLikePost(postID, userID)
	}

	if disliked {
		if err := m.RemoveDislikePost(postID, userID); err != nil {
			return err
		}
	}

	stmt := `INSERT INTO post_likes (post_id, user_id) VALUES (?, ?)`
	_, err = m.DB.Exec(stmt, postID, userID)
	if err != nil {
		return err
	}

	stmt2 := `UPDATE posts SET likes = likes + 1 WHERE id = ?`
	_, err = m.DB.Exec(stmt2, postID)
	return err
}

func (m *ReactionModel) DislikePost(postID, userID int) error {
	liked, err := m.isLiked(postID, userID)
	if err != nil {
		return err
	}
	disliked, err := m.isDisliked(postID, userID)
	if err != nil {
		return err
	}

	if disliked {
		return m.RemoveDislikePost(postID, userID)
	}

	if liked {
		if err := m.RemoveLikePost(postID, userID); err != nil {
			return err
		}
	}

	stmt := `INSERT INTO post_dislikes (post_id, user_id) VALUES (?, ?)`
	_, err = m.DB.Exec(stmt, postID, userID)
	if err != nil {
		return err
	}

	stmt2 := `UPDATE posts SET dislikes = dislikes + 1 WHERE id = ?`
	_, err = m.DB.Exec(stmt2, postID)
	return err
}

func (m *ReactionModel) RemoveLikePost(postID, userID int) error {
	stmt := `DELETE FROM post_likes WHERE post_id = ? AND user_id = ?`
	_, err := m.DB.Exec(stmt, postID, userID)
	if err != nil {
		return err
	}

	stmt2 := `UPDATE posts SET likes = likes - 1 WHERE id = ?`
	_, err = m.DB.Exec(stmt2, postID)
	return err
}

func (m *ReactionModel) RemoveDislikePost(postID, userID int) error {
	stmt := `DELETE FROM post_dislikes WHERE post_id = ? AND user_id = ?`
	_, err := m.DB.Exec(stmt, postID, userID)
	if err != nil {
		return err
	}

	stmt2 := `UPDATE posts SET dislikes = dislikes - 1 WHERE id = ?`
	_, err = m.DB.Exec(stmt2, postID)
	return err
}

func (m *ReactionModel) isCommentLiked(commentID, userID int) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT 1 FROM comment_likes WHERE comment_id = ? AND user_id = ?)`
	err := m.DB.QueryRow(stmt, commentID, userID).Scan(&exists)
	return exists, err
}

func (m *ReactionModel) isCommentDisliked(commentID, userID int) (bool, error) {
	var exists bool
	stmt := `SELECT EXISTS(SELECT 1 FROM comment_dislikes WHERE comment_id = ? AND user_id = ?)`
	err := m.DB.QueryRow(stmt, commentID, userID).Scan(&exists)
	return exists, err
}

func (m *ReactionModel) LikeComment(commentID, userID int) error {
	liked, err := m.isCommentLiked(commentID, userID)
	if err != nil {
		return err
	}
	disliked, err := m.isCommentDisliked(commentID, userID)
	if err != nil {
		return err
	}

	if liked {
		return m.RemoveLikeComment(commentID, userID)
	}

	if disliked {
		if err := m.RemoveDislikeComment(commentID, userID); err != nil {
			return err
		}
	}

	stmt := `INSERT INTO comment_likes (comment_id, user_id) VALUES (?, ?)`
	_, err = m.DB.Exec(stmt, commentID, userID)
	if err != nil {
		return err
	}

	stmt2 := `UPDATE comments SET likes = likes + 1 WHERE id = ?`
	_, err = m.DB.Exec(stmt2, commentID)
	return err
}

func (m *ReactionModel) DislikeComment(commentID, userID int) error {
	liked, err := m.isCommentLiked(commentID, userID)
	if err != nil {
		return err
	}
	disliked, err := m.isCommentDisliked(commentID, userID)
	if err != nil {
		return err
	}

	if disliked {
		return m.RemoveDislikeComment(commentID, userID)
	}

	if liked {
		if err := m.RemoveLikeComment(commentID, userID); err != nil {
			return err
		}
	}

	stmt := `INSERT INTO comment_dislikes (comment_id, user_id) VALUES (?, ?)`
	_, err = m.DB.Exec(stmt, commentID, userID)
	if err != nil {
		return err
	}

	stmt2 := `UPDATE comments SET dislikes = dislikes + 1 WHERE id = ?`
	_, err = m.DB.Exec(stmt2, commentID)
	return err
}
func (m *ReactionModel) RemoveLikeComment(commentID, userID int) error {
	stmt := `DELETE FROM comment_likes WHERE comment_id = ? AND user_id = ?`
	_, err := m.DB.Exec(stmt, commentID, userID)
	if err != nil {
		return err
	}

	stmt2 := `UPDATE comments SET likes = likes - 1 WHERE id = ?`
	_, err = m.DB.Exec(stmt2, commentID)
	return err
}

func (m *ReactionModel) RemoveDislikeComment(commentID, userID int) error {
	stmt := `DELETE FROM comment_dislikes WHERE comment_id = ? AND user_id = ?`
	_, err := m.DB.Exec(stmt, commentID, userID)
	if err != nil {
		return err
	}

	stmt2 := `UPDATE comments SET dislikes = dislikes - 1 WHERE id = ?`
	_, err = m.DB.Exec(stmt2, commentID)
	return err
}
