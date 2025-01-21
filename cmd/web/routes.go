package main

import "net/http"

// Роутер возвращающий сервмукс с роутами нашего приложения
// Переход от mux -> app.routes
func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Регистрация файл-сервера как обработчик для всех URL начинающиеся со static
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Роуты приложения
	mux.Handle("/post/view/", app.sessionManager.LoadAndSave(http.HandlerFunc(app.postView)))
	mux.Handle("/post/create", app.sessionManager.LoadAndSave(http.HandlerFunc(app.postCreateForm)))
	mux.Handle("/", app.sessionManager.LoadAndSave(http.HandlerFunc(app.home)))
	mux.Handle("/user/signup", app.sessionManager.LoadAndSave(http.HandlerFunc(app.userSignup)))
	mux.Handle("/user/login", app.sessionManager.LoadAndSave(http.HandlerFunc(app.userLogin)))
	mux.Handle("user/logout", app.sessionManager.LoadAndSave(http.HandlerFunc(app.userLogout)))

	return app.recoverPanic(app.logRequest(secureHeaders(mux)))
}
