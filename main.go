package main

import (
	"forum/handlers/auth"
	"forum/handlers/post"
	"net/http"
	"time"

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

func handleFunc() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", post.Index)
	mux.HandleFunc("/create", post.CreatePost)
	mux.HandleFunc("/save_post", post.Save_post)
	mux.HandleFunc("/post/", post.Show_post)
	mux.HandleFunc("/auth", auth.Auth)
	mux.HandleFunc("/register", auth.Register)
	mux.HandleFunc("/save_register", auth.Save_register)
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
