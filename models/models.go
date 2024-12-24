package models

type Post struct {
	ID   int
	Name string
	Body string
	Date string
}
type Users struct {
	ID       int
	Name     string
	Email    string
	Password string
}

const Path = "./forum.db"
