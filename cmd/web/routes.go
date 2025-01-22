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
	mux.Handle("/post/view/", http.HandlerFunc(app.postView))
	mux.Handle("/post/create", app.requireAuthentication(http.HandlerFunc(app.postCreateForm)))
	mux.Handle("/", http.HandlerFunc(app.home))
	mux.Handle("/user/signup", http.HandlerFunc(app.userSignup))
	mux.Handle("/user/login", http.HandlerFunc(app.userLogin))
	mux.Handle("/user/logout", http.HandlerFunc(app.userLogout))
	mux.Handle("/user/profile/", http.HandlerFunc(app.profile))
	mux.Handle("/user/profile/changepassword", http.HandlerFunc(app.changePassword))
	return app.recoverPanic(app.logRequest(secureHeaders(mux)))
}
