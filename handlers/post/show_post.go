package post

import (
	"database/sql"
	"forum/models"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

func Show_post(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Регулярное выражение для извлечения ID из URL
	re := regexp.MustCompile(`^/post/([0-9]+)$`)
	matches := re.FindStringSubmatch(r.URL.Path)
	if matches == nil {
		http.NotFound(w, r)
		return
	}

	// Извлекаем ID из URL
	id, err := strconv.Atoi(matches[1])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Открытие шаблонов
	t, err := template.ParseFiles("templates/show_post.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблонов", http.StatusInternalServerError)
		log.Printf("Ошибка загрузки шаблонов: %v", err)
		return
	}

	// Открытие соединения с базой данных
	db, err := sql.Open("sqlite3", models.Path)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		log.Printf("Ошибка открытия базы данных: %v", err)
		return
	}
	defer db.Close()

	// Выборка данных
	var post models.Post
	err = db.QueryRow("SELECT * FROM Posts WHERE ID = ?", id).Scan(&post.ID, &post.Name, &post.Body, &post.Date)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
		log.Printf("Ошибка выполнения запроса: %v", err)
		return
	}

	// Рендеринг шаблона
	err = t.ExecuteTemplate(w, "show_post", post)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка рендеринга шаблона: %v", err)
	}
}
