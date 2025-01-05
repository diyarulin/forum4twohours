package post

import (
	"html/template"
	"log"
	"net/http"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем имя пользователя из cookie
	cookie, err := r.Cookie("userName")
	if err != nil {
		// Если cookie нет, возвращаем ошибку или перенаправляем
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	// Загружаем шаблон и передаем имя пользователя в шаблон
	t, err := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблонов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Передаем имя пользователя в шаблон
	data := map[string]string{
		"UserName": cookie.Value, // Имя пользователя из cookie
	}

	err = t.ExecuteTemplate(w, "create", data)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка рендеринга шаблона: %v", err)
	}
}
