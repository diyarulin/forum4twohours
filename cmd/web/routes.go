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
	mux.Handle("/user/logout", app.requireAuthentication(http.HandlerFunc(app.userLogout)))
	mux.Handle("/user/profile/", app.requireAuthentication(http.HandlerFunc(app.profile)))
	mux.Handle("/user/profile/changepassword", app.requireAuthentication(http.HandlerFunc(app.changePassword)))
	mux.Handle("/post/edit/", app.requireAuthentication(http.HandlerFunc(app.EditPost)))
	mux.Handle("/post/delete/", app.requireAuthentication(http.HandlerFunc(app.DeletePost)))
	mux.Handle("/post/like", app.requireAuthentication(http.HandlerFunc(app.likePost)))
	mux.Handle("/post/dislike", app.requireAuthentication(http.HandlerFunc(app.dislikePost)))
	mux.Handle("/post/remove-like", app.requireAuthentication(http.HandlerFunc(app.removeLikePost)))
	mux.Handle("/post/remove-dislike", app.requireAuthentication(http.HandlerFunc(app.removeDislikePost)))

	mux.Handle("/comment/like", app.requireAuthentication(http.HandlerFunc(app.likeComment)))
	mux.Handle("/comment/dislike", app.requireAuthentication(http.HandlerFunc(app.dislikeComment)))
	mux.Handle("/comment/remove-like", app.requireAuthentication(http.HandlerFunc(app.removeLikeComment)))
	mux.Handle("/comment/remove-dislike", app.requireAuthentication(http.HandlerFunc(app.removeDislikeComment)))
	// Маршруты для комментариев
	mux.Handle("/comments/add", app.requireAuthentication(http.HandlerFunc(app.addComment)))
	mux.Handle("/comment/delete", app.requireAuthentication(http.HandlerFunc(app.deleteComment)))
	mux.Handle("/notifications", app.requireAuthentication(http.HandlerFunc(app.notifications)))
	mux.Handle("/user/googlecallback", http.HandlerFunc(app.googleCallbackHandler))
	mux.Handle("/user/login/google", http.HandlerFunc(app.googleLogin))
	mux.Handle("/user/githubcallback", http.HandlerFunc(app.githubCallbackHandler))
	mux.Handle("/user/login/github", http.HandlerFunc(app.githubLogin))

	mux.Handle("/moderation", app.requireAuthentication(app.requireRole("moderator", http.HandlerFunc(app.moderationPanel))))
	mux.Handle("/post/approve", app.requireRole("moderator", http.HandlerFunc(app.approvePost)))

	// Admin routes
	mux.Handle("/admin/users/promote", app.requireRole("admin", http.HandlerFunc(app.promoteUser)))
	mux.Handle("/admin/users/demote", app.requireRole("admin", http.HandlerFunc(app.demoteUser)))

	mux.Handle("/report/post/", app.requireRole("moderator", http.HandlerFunc(app.reportPost)))
	mux.Handle("/report/answer/", app.requireRole("admin", http.HandlerFunc(app.Answer)))

	mux.Handle("/reports", (http.HandlerFunc(app.viewReports)))
	mux.Handle("/admin/reports", app.requireRole("admin", http.HandlerFunc(app.viewAdminReports)))

	mux.Handle("/admin/categories", app.requireRole("admin", http.HandlerFunc(app.manageCategories)))
	mux.Handle("/admin/categories/add", app.requireRole("admin", http.HandlerFunc(app.addCategory)))
	mux.Handle("/admin/categories/update", app.requireRole("admin", http.HandlerFunc(app.updateCategory)))
	mux.Handle("/admin/categories/delete", app.requireRole("admin", http.HandlerFunc(app.deleteCategory)))

	mux.Handle("/admin/users", app.requireRole("admin", http.HandlerFunc(app.manageUsers)))

	mux.Handle("/user/apply-moderator", app.requireAuthentication(http.HandlerFunc(app.applyForModerator)))
	return app.recoverPanic(app.logRequest(secureHeaders(mux)))
}
