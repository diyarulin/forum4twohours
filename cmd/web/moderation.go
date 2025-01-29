package main

import (
	"net/http"
	"strconv"
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
		pendingPosts, err := app.posts.GetPendingPosts()
		if err != nil {
			app.serverError(w, err)
			return
		}
		data.PendingPosts = pendingPosts // важно это поле

		// Для админов
		if user.Role == "admin" {
			users, err := app.users.GetAllUsers()
			if err != nil {
				app.serverError(w, err)
				return
			}
			data.Users = users
		}
	}

	app.render(w, http.StatusOK, "moderation.html", data)
}

func (app *application) approvePost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.posts.ApprovePost(postID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.flash(w, r, "Post approved successfully!")
	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
}

func (app *application) promoteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.users.PromoteUser(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.flash(w, r, "User promoted to moderator!")
	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
}

func (app *application) demoteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.users.DemoteUser(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.flash(w, r, "User demoted to regular user!")
	http.Redirect(w, r, "/moderation", http.StatusSeeOther)
}
