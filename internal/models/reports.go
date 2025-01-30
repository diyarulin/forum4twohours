// models/reports.go
package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Report struct {
	ID         int
	PostID     int
	ReporterID int
	Reason     string
	CreatedAt  time.Time
	Answer     string
	AdminID    int
	Solved     int
}

type ReportModel struct {
	DB *sql.DB
}

func (m *ReportModel) Get(id int) (*Report, error) {
	stmt := `SELECT id, post_id, reporter_id, reason, created_at, admin_id, answer FROM reports`

	row := m.DB.QueryRow(stmt, id)

	r := &Report{}
	err := row.Scan(&r.ID, &r.PostID, &r.ReporterID, &r.Reason, &r.CreatedAt, &r.AdminID, &r.Answer)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	return r, nil
}

func (m *ReportModel) Create(postID, reporterID int, reason string) error {
	query := `INSERT INTO reports (post_id, reporter_id, reason) VALUES (?, ?, ?)`
	_, err := m.DB.Exec(query, postID, reporterID, reason)
	if err != nil {
		return err
	}
	return nil
}

func (m *ReportModel) Answer(reportID, adminID int, answer string) error {
	var exists bool
	err := m.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM reports WHERE id = ?)", reportID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("ошибка: отчёт с ID %d не найден", reportID)
	}

	query := `UPDATE reports SET admin_id = ?, answer = ?, solved = 1 WHERE id = ?`
	_, err = m.DB.Exec(query, adminID, answer, reportID)
	if err != nil {
		return err
	}

	return nil
}

func (m *ReportModel) GetUnsolved() ([]*Report, error) {
	query := `SELECT * FROM reports WHERE solved = 0`

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*Report
	for rows.Next() {
		r := &Report{}
		err = rows.Scan(&r.ID, &r.PostID, &r.ReporterID, &r.Reason, &r.CreatedAt, &r.Answer, &r.AdminID, &r.Solved)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return reports, nil
}

func (m *ReportModel) GetSolved() error {
	query := `SELECT * FROM reports WHERE solved = 1`
	_, err := m.DB.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

//func (m *ReportModel) GetReported() ([]*Report, error) {
//	query := `SELECT * FROM reports WHERE solved = 1`
//	rows, err := m.DB.Query(query)
//	if err != nil {
//		return nil, err
//	}
//	defer rows.Close()
//	var reports []*Report
//	for rows.Next() {
//		r := &Report{}
//		err := rows.Scan(&r.ID, &r.PostID, &r.ReporterID, &r.Reason, &r.CreatedAt)
//		if err != nil {
//			return nil, err
//		}
//		reports = append(reports, r)
//	}
//	return reports, nil
//}

func (m *ReportModel) GetAll() ([]*Report, error) {
	query := `SELECT id, post_id, reporter_id, reason, created_at, admin_id, answer FROM reports`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*Report
	for rows.Next() {
		r := &Report{}
		err := rows.Scan(&r.ID, &r.PostID, &r.ReporterID, &r.Reason, &r.CreatedAt, &r.AdminID, &r.Answer)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, nil
}
