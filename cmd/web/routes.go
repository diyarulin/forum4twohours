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
	mux.Handle("/post/edit/", http.HandlerFunc(app.EditPost))
	mux.Handle("/post/delete/", http.HandlerFunc(app.DeletePost))
	// Маршруты для комментариев
	mux.Handle("/comments/add", app.requireAuthentication(http.HandlerFunc(app.addComment)))
	mux.Handle("/comments/delete", app.requireAuthentication(http.HandlerFunc(app.deleteComment)))

	mux.Handle("/user/googlecallback", http.HandlerFunc(app.googleCallbackHandler))
	mux.Handle("/user/login/google", http.HandlerFunc(app.googleLogin))
	mux.Handle("/user/githubcallback", http.HandlerFunc(app.githubCallbackHandler))
	mux.Handle("/user/login/github", http.HandlerFunc(app.githubLogin))
	return app.recoverPanic(app.logRequest(secureHeaders(mux)))
}
