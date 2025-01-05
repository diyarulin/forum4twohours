package post

import (
	"database/sql"
	"forum/handlers/auth"
	"forum/models"
	"html/template"
	"log"
	"net/http"
)

// Обработчик для редактирования поста
func EditPost(w http.ResponseWriter, r *http.Request) {
	// Получаем информацию о текущем пользователе из сессии
	user, err := auth.GetSession(r)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
		return
	}

	// Получаем ID поста из параметра URL
	postID := r.URL.Query().Get("id")

	// Подключаемся к базе данных
	db, err := sql.Open("sqlite3", models.Path)
	if err != nil {
		http.Error(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
		log.Printf("Ошибка подключения к базе данных: %v", err)
		return
	}
	defer db.Close()

	// Получаем текущие данные поста
	var post models.Post
	err = db.QueryRow("SELECT ID, Name, Body, Date, Author FROM Posts WHERE ID = ?", postID).Scan(&post.ID, &post.Name, &post.Body, &post.Date, &post.Author)
	if err != nil {
		http.Error(w, "Ошибка получения поста", http.StatusInternalServerError)
		log.Printf("Ошибка получения поста: %v", err)
		return
	}

	// Проверяем, что пост принадлежит текущему пользователю
	if post.Author != user.Name {
		http.Error(w, "Вы не можете редактировать этот пост", http.StatusForbidden)
		return
	}

	// Загружаем шаблон редактирования
	t, err := template.ParseFiles("templates/edit_post.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблонов", http.StatusInternalServerError)
		log.Printf("Ошибка загрузки шаблонов: %v", err)
		return
	}

	// Передаем данные поста в шаблон
	data := map[string]interface{}{
		"Post": post,
	}

	// Рендерим шаблон
	err = t.ExecuteTemplate(w, "edit_post", data)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка рендеринга шаблона: %v", err)
	}
}

// Обработчик для удаления поста
func DeletePost(w http.ResponseWriter, r *http.Request) {
	// Получаем информацию о текущем пользователе из сессии
	user, err := auth.GetSession(r)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
		return
	}

	// Получаем ID поста из параметра URL
	postID := r.URL.Query().Get("id")

	// Подключаемся к базе данных
	db, err := sql.Open("sqlite3", models.Path)
	if err != nil {
		http.Error(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
		log.Printf("Ошибка подключения к базе данных: %v", err)
		return
	}
	defer db.Close()

	// Получаем данные поста
	var post models.Post
	err = db.QueryRow("SELECT ID, Name, Body, Date, Author FROM Posts WHERE ID = ?", postID).Scan(&post.ID, &post.Name, &post.Body, &post.Date, &post.Author)
	if err != nil {
		http.Error(w, "Ошибка получения поста", http.StatusInternalServerError)
		log.Printf("Ошибка получения поста: %v", err)
		return
	}

	// Проверяем, что пост принадлежит текущему пользователю
	if post.Author != user.Name {
		http.Error(w, "Вы не можете удалять этот пост", http.StatusForbidden)
		return
	}

	// Удаляем пост
	stmt, err := db.Prepare("DELETE FROM Posts WHERE ID = ?")
	if err != nil {
		http.Error(w, "Ошибка подготовки запроса", http.StatusInternalServerError)
		log.Printf("Ошибка подготовки запроса: %v", err)
		return
	}

	_, err = stmt.Exec(postID)
	if err != nil {
		http.Error(w, "Ошибка удаления поста", http.StatusInternalServerError)
		log.Printf("Ошибка удаления поста: %v", err)
		return
	}

	// Перенаправляем на профиль
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
