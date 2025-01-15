package models

type Post struct {
	ID       int
	Name     string
	Body     string
	Category string
	Date     string
	Author   string
}
type Users struct {
	ID       int
	Name     string
	Email    string
	Password string
}

const Path = "./forum.db"
