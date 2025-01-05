package post

import (
	"database/sql"
	"fmt"
	"forum/handlers/auth"
	"forum/models"
	"log"
	"net/http"
	"time"
)

func Save_post(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем данные из формы
	postName := r.FormValue("Name") // Имя поста
	body := r.FormValue("Body")     // Тело поста

	// Проверка на пустые данные
	if postName == "" || body == "" {
		fmt.Fprintf(w, "Информация неполная")
		return
	}

	// Получаем текущего пользователя из сессии
	user, err := auth.GetSession(r) // Получаем текущего пользователя
	if err != nil {
		http.Error(w, "Ошибка получения пользователя из сессии", http.StatusUnauthorized)
		return
	}

	// Используем имя пользователя из сессии
	userName := user.Name

	// Генерация текущей даты и времени
	currentTime := time.Now().Format("2006-01-02 15:04:05") // Форматирование времени

	// Открытие соединения с базой данных
	db, err := sql.Open("sqlite3", models.Path)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		log.Printf("Ошибка открытия базы данных: %v", err)
		return
	}
	defer db.Close()

	// Подготовка SQL запроса
	stmt, err := db.Prepare(`INSERT INTO Posts (Name, Body, Date, Author) VALUES (?, ?, ?, ?)`)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		log.Printf("Ошибка подготовки запроса: %v", err)
		return
	}
	defer stmt.Close()

	// Вставка данных в базу данных
	_, err = stmt.Exec(postName, body, currentTime, userName)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		log.Printf("Ошибка вставки данных: %v", err)
		return
	}

	// Перенаправление на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
