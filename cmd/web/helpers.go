package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Print(trace)

	data := &templateData{
		Status:  http.StatusInternalServerError,
		Message: "An unexpected error occurred. Please try again later.",
	}

	app.render(w, http.StatusInternalServerError, "error.html", data)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	data := &templateData{
		Status:  status,
		Message: http.StatusText(status),
	}

	app.render(w, status, "error.html", data)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	// Если это страница ошибки, используем шаблон без навигации
	if status >= 400 {
		err := ts.ExecuteTemplate(buf, "main", data)
		if err != nil {
			app.serverError(w, err)
			return
		}
	} else {
		err := ts.ExecuteTemplate(buf, "base", data)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (app *application) flash(w http.ResponseWriter, r *http.Request, message string) {
	// Сохраняем сообщение во флеш
	http.SetCookie(w, &http.Cookie{
		Name:     "flash_message",
		Value:    message,
		Path:     "/",
		Expires:  time.Now().Add(5 * time.Minute),
		HttpOnly: true,
	})
}

func (app *application) newTemplateData(w http.ResponseWriter, r *http.Request) *templateData {
	// Извлекаем флеш-сообщение из cookie
	flashMessage, err := r.Cookie("flash_message")
	if err != nil {
		flashMessage = nil
	}

	// Удаляем флеш-сообщение из cookie после его использования
	if flashMessage != nil {
		// Используем w для установки cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "flash_message",
			Value:    "",
			Path:     "/",
			Expires:  time.Now().Add(-time.Hour), // Устанавливаем истекший срок
			HttpOnly: true,
		})
	}

	// Передаем флеш-сообщение в шаблон
	var flash string
	if flashMessage != nil {
		flash = flashMessage.Value // Сохраняем текст сообщения
	}
	data := &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           flash, // Передаем флеш-сообщение как строку
		IsAuthenticated: app.isAuthenticated(r),
	}
	if app.isAuthenticated(r) {
		userID, err := app.getCurrentUser(r)
		if err == nil {
			// Получаем количество непрочитанных
			count, _ := app.notificationsModel.GetUnreadCount(userID)
			data = &templateData{
				CurrentYear:         time.Now().Year(),
				Flash:               flash, // Передаем флеш-сообщение как строку
				IsAuthenticated:     app.isAuthenticated(r),
				UnreadNotifications: count,
			}

		}
	}

	return data
}

func (app *application) isAuthenticated(r *http.Request) bool {
	_, err := app.getCurrentUser(r)
	if err != nil {
		return false
	}
	return true
}
func (app *application) methodNotAllowed(w http.ResponseWriter) {
	app.clientError(w, http.StatusMethodNotAllowed)
}
