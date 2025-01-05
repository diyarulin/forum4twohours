package post

import (
	"database/sql"
	"forum/handlers/auth"
	"forum/models"
	"log"
	"net/http"
)

// Обработчик для сохранения изменений в посте
func SaveEditPost(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем информацию о текущем пользователе из сессии
	user, err := auth.GetSession(r)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
		return
	}

	// Получаем данные из формы
	postID := r.FormValue("id")     // ID поста
	postName := r.FormValue("Name") // Название поста
	body := r.FormValue("Body")     // Текст поста

	// Проверка на пустые данные
	if postName == "" || body == "" {
		http.Error(w, "Информация неполная", http.StatusBadRequest)
		return
	}

	// Подключаемся к базе данных
	db, err := sql.Open("sqlite3", models.Path)
	if err != nil {
		http.Error(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
		log.Printf("Ошибка подключения к базе данных: %v", err)
		return
	}
	defer db.Close()

	// Получаем текущий пост
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

	// Обновляем пост в базе данных
	stmt, err := db.Prepare("UPDATE Posts SET Name = ?, Body = ? WHERE ID = ?")
	if err != nil {
		http.Error(w, "Ошибка подготовки запроса", http.StatusInternalServerError)
		log.Printf("Ошибка подготовки запроса: %v", err)
		return
	}

	_, err = stmt.Exec(postName, body, postID)
	if err != nil {
		http.Error(w, "Ошибка сохранения изменений", http.StatusInternalServerError)
		log.Printf("Ошибка сохранения изменений: %v", err)
		return
	}

	// Перенаправляем на страницу профиля
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
