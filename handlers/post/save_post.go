package post

import (
	"database/sql"
	"fmt"
	"forum/models"
	"log"
	"net/http"
)

func Save_post(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	Name := r.FormValue("Name")
	Body := r.FormValue("Body")
	Date := r.FormValue("Date")

	if Name == "" || Body == "" || Date == "" {
		fmt.Fprintf(w, "Information is empty")
	} else {
		// Открытие соединения с базой данных
		db, err := sql.Open("sqlite3", models.Path)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			log.Printf("Ошибка открытия базы данных: %v", err)
			return
		}
		defer db.Close()

		// Использование подготовленного выражения для безопасной вставки данных
		stmt, err := db.Prepare(`INSERT INTO Posts (Name, Body, Date) VALUES (?, ?, ?)`)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			log.Printf("Ошибка подготовки выражения: %v", err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(Name, Body, Date)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			log.Printf("Ошибка вставки данных: %v", err)
			return
		}

		// Перенаправление на главную страницу
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}
