package post

import (
	"database/sql"
	"fmt"
	"forum/models"
	"html/template"
	"log"
	"net/http"
)

var posts = []models.Post{}

// Index обработчик для главной страницы
func Index(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Получаем выбранную категорию из параметров запроса
	selectedCategory := r.URL.Query().Get("Category")

	// Загружаем шаблоны
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	posts = []models.Post{}

	// Открытие соединения с базой данных
	db, err := sql.Open("sqlite3", models.Path)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		log.Printf("Ошибка открытия базы данных: %v", err)
		return
	}
	defer db.Close()

	// Выборка данных
	var query string
	var args []interface{}
	if selectedCategory == "" {
		query = "SELECT ID, Name, Body, Category, Date, Author FROM Posts"
	} else {
		query = "SELECT ID, Name, Body, Category, Date, Author FROM Posts WHERE Category = ?"
		args = append(args, selectedCategory)
	}

	res, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Ошибка получения данных", http.StatusInternalServerError)
		log.Printf("Ошибка выборки данных: %v", err)
		return
	}
	defer res.Close()

	// Обработка результатов выборки
	for res.Next() {
		var post models.Post
		err := res.Scan(&post.ID, &post.Name, &post.Body, &post.Category, &post.Date, &post.Author)
		if err != nil {
			http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
			log.Printf("Ошибка сканирования данных: %v", err)
			return
		}
		posts = append(posts, post)
	}

	// Передаем данные в шаблон
	data := struct {
		Posts            []models.Post
		SelectedCategory string
	}{
		Posts:            posts,
		SelectedCategory: selectedCategory,
	}

	// Рендеринг шаблона
	err = t.ExecuteTemplate(w, "index", data)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка рендеринга шаблона: %v", err)
	}
}
