package main

import (
	"errors"
	"forum/internal/models"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func (app *application) moderationPanel(w http.ResponseWriter, r *http.Request) {
	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	user, err := app.users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(w, r)
	data.User = user

	if user.Role == "moderator" {
		// Получаем посты, ожидающие модерации
		pendingPosts, err := app.posts.GetPendingPosts()
		if err != nil {
			app.serverError(w, err)
			return
		}
		data.PendingPosts = pendingPosts
	}

	if user.Role == "admin" {
		// Получаем отчеты для администратора
		reports, err := app.reports.GetAll()
		if err != nil {
			app.serverError(w, err)
			return
		}
		data.Reports = reports
	}

	app.render(w, http.StatusOK, "moderation.html", data)
}

////
////// MODERATOR
////func (app *application) approvePost(w http.ResponseWriter, r *http.Request) {
////	postID := r.FormValue("post_id")
////	// Логика для одобрения поста
////	err := app.posts.ApprovePost(postID)
////	if err != nil {
////		app.serverError(w, err)
////		return
////	}
////	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
////}
////
////func (app *application) deletePost(w http.ResponseWriter, r *http.Request) {
////	idParam := r.URL.Query().Get("id")
////	var id int
////	var err error
////
////	if idParam != "" {
////		// Если параметр есть, преобразуем его в число
////		id, err = strconv.Atoi(idParam)
////	} else {
////		// Иначе пытаемся извлечь ID из пути
////		path := strings.TrimPrefix(r.URL.Path, "/post/delete/")
////		id, err = strconv.Atoi(path)
////	}
////	path, err := app.posts.DeletePost(id)
////	if err != nil {
////		app.serverError(w, err)
////	}
////	if path != "" {
////		err = os.Remove("./ui/static/upload/" + path)
////		if err != nil {
////			app.serverError(w, err)
////			return
////		}
////	}
////	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
////}
////
////func (app *application) reportPost(w http.ResponseWriter, r *http.Request) {
////	postID, _ := strconv.Atoi(r.FormValue("post_id"))
////	// Logic to report the post to admin
////}
////
////// ADMINISTRATOR
////func (app *application) promoteUser(w http.ResponseWriter, r *http.Request) {
////	userID, _ := strconv.Atoi(r.FormValue("user_id"))
////	// Логика для повышения пользователя до модератора
////	err := app.users.PromoteUser(userID)
////	if err != nil {
////		app.serverError(w, err)
////		return
////	}
////	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
////}
////
////func (app *application) demoteUser(w http.ResponseWriter, r *http.Request) {
////	userID := r.FormValue("user_id")
////	// Логика для понижения пользователя
////	err := app.users.DemoteUser(userID)
////	if err != nil {
////		app.serverError(w, err)
////		return
////	}
////	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
////}
////
////func (app *application) manageCategories(w http.ResponseWriter, r *http.Request) {
////	// Logic to add or delete categories
////}
////
////func (app *application) handleReport(w http.ResponseWriter, r *http.Request) {
////	reportID := r.FormValue("report_id")
////	// Logic to handle the report
////}
