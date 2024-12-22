package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Posts struct {
	ID   int
	Name string
	Body string
	Date string
	User string
}

func index(w http.ResponseWriter, r *http.Request) {
	var posts = []Posts{}
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	// Открытие соединения с базой данных
	path := "./forum.db"
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		log.Printf("Ошибка открытия базы данных: %v", err)
		return
	}
	defer db.Close()
	// Выборка данных
	res, err := db.Query("SELECT * FROM Posts")
	if err != nil {
		log.Fatalf("Error selecting data: %v", err)
	}

	for res.Next() {
		var post Posts
		// Укажите все поля структуры для сканирования данных
		err := res.Scan(&post.ID, &post.Name, &post.Body, &post.Date, &post.User)
		if err != nil {
			log.Fatalf("Error scanning data: %v", err)
		}
		posts = append(posts, post)

	}
	t.ExecuteTemplate(w, "index", posts)
}
func create(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	t.ExecuteTemplate(w, "create", nil)
}
func save_post(w http.ResponseWriter, r *http.Request) {
	Name := r.FormValue("Name")
	Body := r.FormValue("Body")
	Date := r.FormValue("Date")
	User := r.FormValue("User")
	path := "./forum.db"
	if Name == "" || Body == "" || Date == "" || User == "" {
		fmt.Fprintf(w, "Information is empty")
	} else {
		// Открытие соединения с базой данных
		db, err := sql.Open("sqlite3", path)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			log.Printf("Ошибка открытия базы данных: %v", err)
			return
		}
		defer db.Close()

		// Использование подготовленного выражения для безопасной вставки данных
		stmt, err := db.Prepare(`INSERT INTO Posts (Name, Body, Date, User) VALUES (?, ?, ?, ?)`)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			log.Printf("Ошибка подготовки выражения: %v", err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(Name, Body, Date, User)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			log.Printf("Ошибка вставки данных: %v", err)
			return
		}

		// Перенаправление на главную страницу
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}

func handleFunc() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.HandleFunc("/", index)
	http.HandleFunc("/create", create)
	http.HandleFunc("/save_post", save_post)
	http.ListenAndServe(":8080", nil)
}

func main() {
	handleFunc()
}
