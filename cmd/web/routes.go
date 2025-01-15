package main

import "net/http"

// Роутер возвращающий сервмукс с роутами нашего приложения
// Переход от mux -> app.routes
func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()

	// Регистрация файл-сервера как обработчик для всех URL начинающиеся со static
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Роуты приложения
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/post/view", app.postView)
	mux.HandleFunc("/post/create", app.postCreate)

	return mux
}
