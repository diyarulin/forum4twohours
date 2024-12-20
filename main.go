package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type Posts struct {
	ID   int
	Name string
	Body string
	Date string
	User []string
}

var posts = Posts{
	ID:   1,
	Name: "Welcome to My Blog",
	Body: "This is the first post on my new blog. I will share my thoughts on various topics.",
	Date: "2024-10-19",
	User: []string{"John Doe", "Didar"},
}

// {
// 	ID:   2,
// 	Name: "Understanding Go Structs",
// 	Body: "In this post, we will dive into how structs work in Go and how to use them efficiently.",
// 	Date: "2024-10-20",
// 	User: "Jane Smith",
// },
// {
// 	ID:   3,
// 	Name: "Working with HTTP in Go",
// 	Body: "Let's explore how to build a simple web server in Go using the net/http package.",
// 	Date: "2024-10-21",
// 	User: "DevGuru",
// },
// }

func (p *Posts) getAllInfo() string {
	return fmt.Sprintf("id is: %d, title is: %s, context is: %s,  date is: %s, username is: %s", p.ID, p.Name, p.Body, p.Date, p.User)
}
func (p *Posts) editName(newTitle string) {
	p.Name = newTitle
}
func home_page(w http.ResponseWriter, r *http.Request) {
	// Post := Posts{1, "How to write forum", "1st step, 2nd step", "David", "2017-08-31"}
	// // Post.editTitle("123")
	// fmt.Fprint(w, `<h1>Main Text</h1>
	// <b>Main</b>`)
	tmpl, _ := template.ParseFiles("templates/home_page.html")
	tmpl.Execute(w, posts)
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
