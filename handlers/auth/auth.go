package auth

import (
	"database/sql"
	"forum/models"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Auth отвечает за обработку аутентификации
func Auth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodGet {
		t, err := template.ParseFiles("templates/auth.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки шаблонов", http.StatusInternalServerError)
			log.Printf("Ошибка загрузки шаблонов: %v", err)
			return
		}

		err = t.ExecuteTemplate(w, "auth", nil)
		if err != nil {
			http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
			log.Printf("Ошибка рендеринга шаблона: %v", err)
		}
		return
	}

	// Обработка POST-запроса
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Очищаем пробелы в начале и в конце строки
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)

	log.Printf("Email: %s, Password: %s\n", email, password)

	// Проверка на пустые значения
	if email == "" || password == "" {
		t, err := template.ParseFiles("templates/auth.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки шаблонов", http.StatusInternalServerError)
			log.Printf("Ошибка загрузки шаблонов: %v", err)
			return
		}

		data := map[string]string{
			"ErrorMessage": "Email или пароль не могут быть пустыми",
		}

		t.ExecuteTemplate(w, "auth", data)
		return
	}

	// Проверка формата email
	emailRegex := `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	if !re.MatchString(email) {
		http.Error(w, "Некорректный email", http.StatusBadRequest)
		return
	}

	// Подключение к базе данных
	db, err := sql.Open("sqlite3", models.Path)
	if err != nil {
		http.Error(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
		log.Printf("Ошибка подключения к базе данных: %v", err)
		return
	}
	defer db.Close()

	var dbPassword, userName string
	err = db.QueryRow("SELECT Name, Password FROM Users WHERE Email = ?", email).Scan(&userName, &dbPassword)
	log.Printf("DBName: %s, DBPassword: %s\n", userName, dbPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		} else {
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			log.Printf("Ошибка запроса к базе данных: %v", err)
		}
		return
	}

	// Сравнение паролей с использованием bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password))
	if err != nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	// Установка cookie с именем session
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    email, // Используем email для идентификации пользователя
		HttpOnly: true,
		Path:     "/",
	})

	// Перенаправление на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
