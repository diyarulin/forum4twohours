package post

import (
	"database/sql"
	"fmt"
	"forum/models"
	"html/template"
	"log"
	"net/http"
)

var posts = []models.Posts{}

func Index(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	posts = []models.Posts{}
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
		var post models.Posts
		err := res.Scan(&post.ID, &post.Name, &post.Body, &post.Date, &post.User)
		if err != nil {
			log.Fatalf("Error scanning data: %v", err)
		}
		posts = append(posts, post)

	}
	t.ExecuteTemplate(w, "index", posts)
}
