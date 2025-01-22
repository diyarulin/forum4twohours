package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

// The serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Print(trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the user. We'll use this later in the book to send responses like 400 "Bad
// Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// For consistency, we'll also implement a notFound helper. This is simply a
// convenience wrapper around clientError which sends a 404 Not Found response to
// the user.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	// Retrieve the appropriate template set from the cache based on the page
	// name (like 'home.html'). If no entry exists in the cache with the
	// provided name, then create a new error and call the serverError() helper
	// method that we made earlier and return.
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Write out the provided HTTP status code ('200 OK', '400 Bad Request'
	// etc.).
	w.WriteHeader(status)

	buf.WriteTo(w)
	// Execute the template set and write the response body. Again, if there
	// is any error we call the the serverError() helper.
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

	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           flash, // Передаем флеш-сообщение как строку
		IsAuthenticated: app.isAuthenticated(r),
	}
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
