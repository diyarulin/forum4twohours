package auth

import (
	"html/template"
	"net/http"
)

// Register отображает страницу регистрации
func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	t, err := template.ParseFiles("templates/register.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблонов", http.StatusInternalServerError)
		return
	}
	err = t.ExecuteTemplate(w, "register", nil)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
	}
}
