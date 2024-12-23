package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

// TODO
// Сделать таблицу в forum.db по структуре UserDetails
// Добавить регистрацию, помимо входа, т.к. пока не с чем сравнивать логин и пароль
// Добавить комментарии, лайки
// Добавить возможность редактирования и удаления постов
// Поиграть с оформлением
// Подключить UUID
//
//	Разбить все по пакетам
//	Добавить докер
//	Добавить возможность добавлять картинки
//	Связать логин и имя пользователя, чтобы отображалось в посте и связать в базе данных
//
// Время должно само отображаться в соответствии со временем публикации поста, а не вводиться вручную
type Posts struct {
	ID   int
	Name string
	Body string
	Date string
	User string
}
type UserDetails struct {
	Login         string
	Password      string
	Success       bool
	StorageAccess string
}

var posts = []Posts{}
var showPost = []Posts{}
var path = "./forum.db"

func index(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	posts = []Posts{}
	// Открытие соединения с базой данных
	path := "./forum.db"
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		log.Printf("Ошибка открытия базы данных: %v", err)
		return
	}
	defer db.Close()
	// Выборка данных
	res, err := db.Query("SELECT * FROM Posts")
	if err != nil {
		log.Fatalf("Error selecting data: %v", err)
	}
	for res.Next() {
		var post Posts
		err := res.Scan(&post.ID, &post.Name, &post.Body, &post.Date, &post.User)
		if err != nil {
			log.Fatalf("Error scanning data: %v", err)
		}
		posts = append(posts, post)

	}
	t.ExecuteTemplate(w, "index", posts)
}
func create(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	t, err := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	t.ExecuteTemplate(w, "create", nil)
}
func save_post(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	Name := r.FormValue("Name")
	Body := r.FormValue("Body")
	Date := r.FormValue("Date")
	User := r.FormValue("User")

	if Name == "" || Body == "" || Date == "" || User == "" {
		fmt.Fprintf(w, "Information is empty")
	} else {
		// Открытие соединения с базой данных
		db, err := sql.Open("sqlite3", path)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			log.Printf("Ошибка открытия базы данных: %v", err)
			return
		}
		defer db.Close()

		// Использование подготовленного выражения для безопасной вставки данных
		stmt, err := db.Prepare(`INSERT INTO Posts (Name, Body, Date, User) VALUES (?, ?, ?, ?)`)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			log.Printf("Ошибка подготовки выражения: %v", err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(Name, Body, Date, User)
		if err != nil {
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			log.Printf("Ошибка вставки данных: %v", err)
			return
		}

		// Перенаправление на главную страницу
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

}
func show_post(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Регулярное выражение для извлечения ID из URL
	re := regexp.MustCompile(`^/post/([0-9]+)$`)
	matches := re.FindStringSubmatch(r.URL.Path)
	if matches == nil {
		http.NotFound(w, r)
		return
	}

	// Извлекаем ID из URL
	id, err := strconv.Atoi(matches[1])
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Открытие шаблонов
	t, err := template.ParseFiles("templates/show_post.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		http.Error(w, "Ошибка загрузки шаблонов", http.StatusInternalServerError)
		log.Printf("Ошибка загрузки шаблонов: %v", err)
		return
	}

	// Открытие соединения с базой данных
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		log.Printf("Ошибка открытия базы данных: %v", err)
		return
	}
	defer db.Close()

	// Выборка данных
	var post Posts
	err = db.QueryRow("SELECT * FROM Posts WHERE ID = ?", id).Scan(&post.ID, &post.Name, &post.Body, &post.Date, &post.User)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Ошибка выполнения запроса", http.StatusInternalServerError)
		log.Printf("Ошибка выполнения запроса: %v", err)
		return
	}

	// Рендеринг шаблона
	err = t.ExecuteTemplate(w, "show_post", post)
	if err != nil {
		http.Error(w, "Ошибка рендеринга шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка рендеринга шаблона: %v", err)
	}
}
func auth(w http.ResponseWriter, r *http.Request) {
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
	data := UserDetails{
		Login:    r.FormValue("login"),
		Password: r.FormValue("password"),
	}

	// Здесь может быть логика для проверки пользователя.
	// Например:
	if data.Login == "admin" && data.Password == "password" {
		data.Success = true
		data.StorageAccess = "Full"
	} else {
		data.Success = false
		data.StorageAccess = "None"
	}

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
func handleFunc() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/create", create)
	mux.HandleFunc("/save_post", save_post)
	mux.HandleFunc("/post/", show_post)
	mux.HandleFunc("/auth/", auth)
	http.Handle("/", mux)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.ListenAndServe(":8080", mux)
}

func main() {
	handleFunc()
}
