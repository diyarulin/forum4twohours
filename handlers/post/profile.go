package post

import (
	"database/sql"
	"forum/handlers/auth"
	"forum/models"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Обработчик для отображения профиля пользователя
func Profile(w http.ResponseWriter, r *http.Request) {
	// Получаем информацию о текущем пользователе из сессии
	user, err := auth.GetSession(r)
	if err != nil {
		// Перенаправление на главную страницу
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	// Получаем все посты пользователя
	db, err := sql.Open("sqlite3", models.Path)
	if err != nil {
		http.Error(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
		log.Printf("Ошибка подключения к базе данных: %v", err)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT ID, Name, Body, Date FROM Posts WHERE Author = ?", user.Name)
	if err != nil {
		http.Error(w, "Ошибка получения постов", http.StatusInternalServerError)
		log.Printf("Ошибка получения постов: %v", err)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Name, &post.Body, &post.Date); err != nil {
			http.Error(w, "Ошибка получения данных поста", http.StatusInternalServerError)
			log.Printf("Ошибка получения данных поста: %v", err)
			return
		}
		posts = append(posts, post)
	}

	// Загружаем шаблон профиля
	t, err := template.ParseFiles("templates/profile.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблонов", http.StatusInternalServerError)
		log.Printf("Ошибка загрузки шаблонов: %v", err)
		return
	}

	// Передаем данные пользователя и его посты в шаблон
	data := map[string]interface{}{
		"User":  user,
		"Posts": posts,
	}

	// Рендерим шаблон
	err = t.ExecuteTemplate(w, "profile", data)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка рендеринга шаблона: %v", err)
	}
}

// Обработчик для изменения пароля
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем информацию о текущем пользователе из сессии
	user, err := auth.GetSession(r)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
		return
	}

	// Получаем текущий и новый пароль из формы
	currentPassword := r.FormValue("currentPassword")
	newPassword := r.FormValue("newPassword")
	confirmPassword := r.FormValue("confirmPassword")

	// Проверка на пустые значения
	if currentPassword == "" || newPassword == "" || confirmPassword == "" {
		http.Error(w, "Все поля должны быть заполнены", http.StatusBadRequest)
		return
	}

	// Проверка на совпадение нового пароля с подтверждением
	if newPassword != confirmPassword {
		http.Error(w, "Пароли не совпадают", http.StatusBadRequest)
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

	// Получаем текущий пароль из базы данных
	var dbPassword string
	err = db.QueryRow("SELECT Password FROM Users WHERE Email = ?", user.Email).Scan(&dbPassword)
	if err != nil {
		http.Error(w, "Ошибка поиска пользователя", http.StatusInternalServerError)
		log.Printf("Ошибка запроса к базе данных: %v", err)
		return
	}

	// Сравниваем текущий пароль с тем, что хранится в базе данных
	err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(currentPassword))
	if err != nil {
		http.Error(w, "Неверный текущий пароль", http.StatusUnauthorized)
		return
	}

	// Хэшируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка хэширования пароля", http.StatusInternalServerError)
		log.Printf("Ошибка хэширования пароля: %v", err)
		return
	}

	// Обновляем пароль в базе данных
	_, err = db.Exec("UPDATE Users SET Password = ? WHERE Email = ?", hashedPassword, user.Email)
	if err != nil {
		http.Error(w, "Ошибка обновления пароля", http.StatusInternalServerError)
		log.Printf("Ошибка обновления пароля: %v", err)
		return
	}

	// Перенаправление на страницу профиля
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
