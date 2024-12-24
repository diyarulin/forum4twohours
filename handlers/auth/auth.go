package auth

import (
	"forum/models"
	"html/template"
	"log"
	"net/http"
)

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

	// Для POST-запросов
	data := models.Users{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	// Здесь может быть логика для проверки пользователя.
	// Например:
	// if data.Login == "admin" && data.Password == "password" {
	// 	data.Success = true
	// 	data.StorageAccess = "Full"
	// } else {
	// 	data.Success = false
	// 	data.StorageAccess = "None"
	// }

	// Загрузка шаблонов
	t, err := template.ParseFiles("templates/auth.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблонов", http.StatusInternalServerError)
		log.Printf("Ошибка загрузки шаблонов: %v", err)
		return
	}

	// Рендеринг шаблона с данными
	err = t.ExecuteTemplate(w, "auth", data)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка рендеринга шаблона: %v", err)
	}
}
