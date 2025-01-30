// models/reports.go
package models

import (
	"database/sql"
	"time"
)

type Report struct {
	ID         int
	PostID     int
	ReporterID int
	Reason     string
	CreatedAt  time.Time
}

type ReportModel struct {
	DB *sql.DB
}

func (m *ReportModel) Create(postID, reporterID int, reason string) error {
	query := `INSERT INTO reports (post_id, reporter_id, reason) VALUES (?, ?, ?)`
	_, err := m.DB.Exec(query, postID, reporterID, reason)
	return err
}

func (m *ReportModel) GetAll() ([]*Report, error) {
	query := `SELECT id, post_id, reporter_id, reason, created_at FROM reports`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*Report
	for rows.Next() {
		r := &Report{}
		err := rows.Scan(&r.ID, &r.PostID, &r.ReporterID, &r.Reason, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, nil
}