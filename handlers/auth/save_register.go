package auth

import (
	"database/sql"
	"fmt"
	"forum/models"
	"log"
	"net/http"
)

func Save_register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm() // Парсим данные формы
	Name := r.FormValue("Name")
	Email := r.FormValue("Email")
	Password := r.FormValue("Password")

	fmt.Printf("Name: %s, Email: %s, Password: %s\n", Name, Email, Password)

	if Name == "" || Email == "" || Password == "" {
		fmt.Fprintf(w, "Information is empty")
		return
	}

	db, err := sql.Open("sqlite3", models.Path)
	if err != nil {
		http.Error(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
		log.Printf("Ошибка открытия базы данных: %v", err)
		return
	}
	defer db.Close()

	stmt, err := db.Prepare(`INSERT INTO Users (Name, Email, Password) VALUES (?, ?, ?)`)
	if err != nil {
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		log.Printf("Ошибка подготовки выражения: %v", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(Name, Email, Password)
	if err != nil {
		http.Error(w, "Ошибка при вставке данных", http.StatusInternalServerError)
		log.Printf("Ошибка вставки данных: %v", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
