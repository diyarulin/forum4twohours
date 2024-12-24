package models

type Posts struct {
	ID   int
	Name string
	Body string
	Date string
	User string
}
type UserDetails struct {
	Login         string
	Password      string
	Success       bool
	StorageAccess string
}

const Path = "./forum.db"

var showPost = []Posts{}
