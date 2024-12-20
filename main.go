package main

import (
	"fmt"
	"net/http"
)

type Posts struct {
	id       uint16
	title    string
	context  string
	username string
	date     string
}

func (p *Posts) getAllInfo() string {
	return fmt.Sprintf("id is: %d, title is: %s, context is: %s, username is: %s, date is: %s", p.id, p.title, p.context, p.username, p.date)
}
func (p *Posts) editTitle(newTitle string) {
	p.title = newTitle
}
func home_page(w http.ResponseWriter, r *http.Request) {
	Post := Posts{1, "How to write forum", "1st step, 2nd step", "David", "2017-08-31"}
	Post.editTitle("123")
	fmt.Fprint(w, Post.getAllInfo())
}
func about_page(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "About us")
}
func handleRequest() {
	http.HandleFunc("/", home_page)
	http.HandleFunc("/about/", about_page)
	http.ListenAndServe(":8080", nil)
}
func main() {
	handleRequest()
}
