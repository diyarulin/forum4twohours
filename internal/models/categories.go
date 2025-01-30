package models

import (
	"database/sql"
	"strings"
)

type Category struct {
	ID   int
	Name string
}

type CategoryModel struct {
	DB *sql.DB
}

func (m *CategoryModel) GetAll() ([]*Category, error) {
	stmt := `SELECT id, name FROM categories`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		c := &Category{}
		err = rows.Scan(&c.ID, &c.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}
func (m *CategoryModel) Insert(name string) error {
	stmt := `INSERT INTO categories (name) VALUES (?)`
	_, err := m.DB.Exec(stmt, name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateCategory
		}
		return err
	}
	return nil
}

func (m *CategoryModel) Update(id int, newName string) error {
	stmt := `UPDATE categories SET name = ? WHERE id = ?`
	_, err := m.DB.Exec(stmt, newName, id)
	return err
}

func (m *CategoryModel) Delete(id int) error {
	stmt := `DELETE FROM categories WHERE id = ?`
	_, err := m.DB.Exec(stmt, id)
	return err
}
