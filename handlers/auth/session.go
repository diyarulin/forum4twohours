package auth

import (
	"database/sql"
	"forum/models"
	"log"
	"net/http"
	"time"
)

func SetSession(w http.ResponseWriter, r *http.Request, user *models.Users) {
	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{
		Name:     "session",
		Value:    user.Email, // Используем email для идентификации пользователя
		Expires:  expiration,
		HttpOnly: true, // Защищает cookie от доступа через JavaScript
		Path:     "/",  // Обеспечивает доступ ко всем путям сайта
	}
	http.SetCookie(w, &cookie)
	log.Printf("Сессия установлена для пользователя: %s", user.Email)

	// Проверка, что cookie действительно установлена
	cookies := r.Cookies()
	for _, c := range cookies {
		if c.Name == "session" {
			log.Printf("Cookie сессии установлена: %s", c.Value)
			break
		}
	}
}

// Получить текущего пользователя из сессии
func GetSession(r *http.Request) (*models.Users, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		log.Printf("Не удалось получить cookie: %v", err)
		log.Printf("Все cookies в запросе: %v", r.Cookies()) // Логируем все cookies в запросе
		return nil, err
	}

	// Получаем информацию о пользователе из базы данных по email из сессии
	db, err := sql.Open("sqlite3", models.Path)
	if err != nil {
		log.Printf("Ошибка подключения к базе данных: %v", err)
		return nil, err
	}
	defer db.Close()

	var user models.Users
	err = db.QueryRow("SELECT ID, Name, Email, Password FROM Users WHERE Email = ?", cookie.Value).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		log.Printf("Ошибка запроса пользователя с email %s: %v", cookie.Value, err)
		return nil, err
	}

	log.Printf("Пользователь найден: %v", user)
	return &user, nil
}
