package main

import (
	"forum/handlers/auth"
	"forum/handlers/post"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TODO
// 1. Связать auth и Register в одну вкладку и связать их логику. Сделать как на обычных сайтах типа sign up ...
// 2. Добавить комментарии, лайки и в профиле чтобы можно было отслеживать их
// 3. Поиграть с оформлением. Чтобы норм масштабировалось все
// 4. Подключить UUID
// 5. Добавить докер
// 6. Добавить возможность добавлять картинки
// 7. Добавить категории на сайте, в базе данных
// 8. Сделать профиль слева в верхнем углу с аватаром и чтобы никнейм был виден
// 9. Вытащить стили из ссылки в хедере и футере и добавить их в отдельный css файл
// 10. Проверить на ошибки код, сделать тесты.
// 11. Сделать отдельный элемент в структуре и таблице Posts по типу анонса, чтобы на главной был не весь текст поста , а краткое описание

func handleFunc() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", post.Index)
	mux.HandleFunc("/create", post.CreatePost)
	mux.HandleFunc("/save-post", post.Save_post)

	mux.HandleFunc("/post/", post.Show_post)
	mux.HandleFunc("/auth", auth.Auth)
	mux.HandleFunc("/register", auth.Register)
	mux.HandleFunc("/save_register", auth.Save_register)
	mux.HandleFunc("/profile", post.Profile)
	mux.HandleFunc("/change-password", post.ChangePassword)
	mux.HandleFunc("/delete_post/", post.DeletePost)
	mux.HandleFunc("/edit_post/", post.EditPost)
	mux.HandleFunc("/save_edit_post", post.SaveEditPost)

	http.Handle("/", mux)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	server.ListenAndServe()
}

func main() {
	handleFunc()
}
