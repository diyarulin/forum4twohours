package auth

import (
	"database/sql"
	"forum/models"
	"html/template"
	"log"
	"net/http"
)

// Auth отвечает за обработку аутентификации
func Auth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodGet {
		// Загрузка шаблонов
		t, err := template.ParseFiles("templates/auth.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки шаблонов", http.StatusInternalServerError)
			log.Printf("Ошибка загрузки шаблонов: %v", err)
			return
		}

		// Рендеринг шаблона
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

	log.Printf("Email: %s, Password: %s\n", email, password) // Логируем значения
	if email == "" || password == "" {
		t, err := template.ParseFiles("templates/auth.html", "templates/header.html", "templates/footer.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки шаблонов", http.StatusInternalServerError)
			log.Printf("Ошибка загрузки шаблонов: %v", err)
			return
		}

		// Передаем сообщение об ошибке в шаблон
		data := map[string]string{
			"ErrorMessage": "Email или пароль не могут быть пустыми",
		}

		t.ExecuteTemplate(w, "auth", data)
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

	// Проверка email и пароля
	var dbPassword string
	err = db.QueryRow("SELECT Password FROM Users WHERE Email = ?", email).Scan(&dbPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если пользователь не найден
			http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		} else {
			// Другие ошибки
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			log.Printf("Ошибка запроса к базе данных: %v", err)
		}
		return
	}

	// Сравнение паролей
	if password != dbPassword {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	// Если все успешно, перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
