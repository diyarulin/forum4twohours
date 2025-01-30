package main

import (
	"forum/internal/validator"
	"net/http"
	"strconv"
	"strings"
)

type reportCreateForm struct {
	ReportID   int
	PostID     int
	Reason     string
	ReporterID int
	Answer     string
	AdminID    int
	validator.Validator
}

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
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
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
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

// handlers.go
func (app *application) reportPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		path := strings.TrimPrefix(r.URL.Path, "/report/post/")
		postId, err := strconv.Atoi(path)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		data := app.newTemplateData(w, r)
		data.Form = &reportCreateForm{
			PostID: postId,
			Validator: validator.Validator{
				FieldErrors: map[string]string{},
			},
		}
		app.render(w, http.StatusOK, "report.html", data)
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/report/post/")
		postId, err := strconv.Atoi(path)

		if err != nil {
			app.infoLog.Printf("ERROR123123")

			app.clientError(w, http.StatusBadRequest)
			return
		}
		userID, err := app.getCurrentUser(r)
		if err != nil {
			app.clientError(w, http.StatusUnauthorized)
			return
		}

		reason := r.FormValue("reason")
		if reason == "" {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		form := reportCreateForm{
			PostID:     postId,
			Reason:     reason,
			ReporterID: userID,
		}
		form.CheckField(validator.NotBlank(form.Reason), "title", "This field cannot be blank")
		form.CheckField(validator.MaxChars(form.Reason, 50), "title", "This field cannot be longer than 100 characters")
		if !form.Valid() {
			data := app.newTemplateData(w, r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "report.html", data)
			return
		}
		post, err := app.posts.Get(postId)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		if post == nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		err = app.reports.Create(postId, userID, reason)
		if err != nil {
			app.serverError(w, err)
			return
		}

		app.flash(w, r, "Report submitted successfully!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	app.clientError(w, http.StatusMethodNotAllowed)
}

func (app *application) Answer(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		path := strings.TrimPrefix(r.URL.Path, "/report/answer/")
		reportId, err := strconv.Atoi(path)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		data := app.newTemplateData(w, r)
		data.Form = &reportCreateForm{
			ReportID: reportId,
			Validator: validator.Validator{
				FieldErrors: map[string]string{},
			},
		}
		app.render(w, http.StatusOK, "answer.html", data)
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/report/answer/")
		reportId, err := strconv.Atoi(path)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		userID, err := app.getCurrentUser(r)
		if err != nil {
			app.clientError(w, http.StatusUnauthorized)
			return
		}

		answer := r.FormValue("answer")
		if answer == "" {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		form := reportCreateForm{
			ReportID: reportId,
			Answer:   answer,
			AdminID:  userID,
		}
		form.CheckField(validator.NotBlank(form.Answer), "answer", "This field cannot be blank")
		form.CheckField(validator.MaxChars(form.Answer, 50), "answer", "This field cannot be longer than 100 characters")
		if !form.Valid() {
			data := app.newTemplateData(w, r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "answer.html", data)
			return
		}

		err = app.reports.Answer(reportId, userID, answer)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		app.flash(w, r, "Report submitted successfully!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	app.clientError(w, http.StatusMethodNotAllowed)
}

func (app *application) viewReports(w http.ResponseWriter, r *http.Request) {
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

	if user.Role != "admin" && user.Role != "moderator" {
		app.clientError(w, http.StatusForbidden)
		return
	}

	reports, err := app.reports.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(w, r)
	data.Reports = reports
	app.render(w, http.StatusOK, "reports.html", data)
}

func (app *application) viewAdminReports(w http.ResponseWriter, r *http.Request) {
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

	if user.Role != "admin" {
		app.clientError(w, http.StatusForbidden)
		return
	}

	reports, err := app.reports.GetUnsolved()
	if err != nil {
		app.serverError(w, err)
	}

	data := app.newTemplateData(w, r)
	data.Reports = reports
	app.render(w, http.StatusOK, "adminreports.html", data)
}

func (app *application) applyForModerator(w http.ResponseWriter, r *http.Request) {
	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err = app.users.ApplyForModerator(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.flash(w, r, "Your request to become a moderator has been submitted!")
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}
