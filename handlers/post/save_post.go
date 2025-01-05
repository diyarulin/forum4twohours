package post

import (
	"database/sql"
	"fmt"
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
	userName := r.FormValue("UserName") // Имя пользователя
	postName := r.FormValue("Name")     // Имя поста
	body := r.FormValue("Body")

	if userName == "" || postName == "" || body == "" {
		fmt.Fprintf(w, "Information is empty")
		return
	}

	// Генерация текущей даты и времени в строковом формате
	currentTime := time.Now().Format("2006-01-02 15:04:05") // Форматируем как строку

	// Открытие соединения с базой данных
	db, err := sql.Open("sqlite3", models.Path)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		log.Printf("Ошибка открытия базы данных: %v", err)
		return
	}
	defer db.Close()

	// Использование подготовленного выражения для безопасной вставки данных
	stmt, err := db.Prepare(`INSERT INTO Posts (Name, Body, Date, Author) VALUES (?, ?, ?, ?)`)

	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		log.Printf("Ошибка подготовки выражения: %v", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(postName, body, currentTime, userName) // Передаем дату как строку
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		log.Printf("Ошибка вставки данных: %v", err)
		return
	}

	// Перенаправление на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
