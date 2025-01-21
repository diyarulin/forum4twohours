package auth

import (
	"database/sql"
	"fmt"
	"forum/models"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Save_register сохраняет данные о пользователе в базе данных с хешированием пароля
func Save_register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm() // Парсим данные формы
	Name := r.FormValue("Name")
	Email := r.FormValue("Email")
	Password := r.FormValue("Password")

	// Логируем полученные данные
	fmt.Printf("Name: %s, Email: %s, Password: %s\n", Name, Email, Password)

	// Проверка на пустые поля
	if Name == "" || Email == "" || Password == "" {
		http.Error(w, "Все поля должны быть заполнены", http.StatusBadRequest)
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка хеширования пароля", http.StatusInternalServerError)
		log.Printf("Ошибка хеширования пароля: %v", err)
		return
	}

	// Подключение к базе данных
	db, err := sql.Open("sqlite3", models.Path)
	if err != nil {
		http.Error(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
		log.Printf("Ошибка открытия базы данных: %v", err)
		return
	}
	defer db.Close()

	// Подготовка SQL-запроса для вставки нового пользователя
	stmt, err := db.Prepare(`INSERT INTO Users (Name, Email, Password) VALUES (?, ?, ?)`)
	if err != nil {
		http.Error(w, "Ошибка подготовки SQL-запроса", http.StatusInternalServerError)
		log.Printf("Ошибка подготовки SQL-запроса: %v", err)
		return
	}
	defer stmt.Close()

	// Вставка нового пользователя в базу данных
	_, err = stmt.Exec(Name, Email, string(hashedPassword))
	if err != nil {
		http.Error(w, "Ошибка при вставке данных", http.StatusInternalServerError)
		log.Printf("Ошибка вставки данных: %v", err)
		return
	}

	// Перенаправление на главную страницу после успешной регистрации
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
