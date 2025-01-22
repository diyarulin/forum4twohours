package main

import (
	"errors"
	"fmt"
	"forum/internal/models"
	"forum/internal/validator"
	"net/http"
	"strconv"
	"strings"
)

type postCreateForm struct {
	Title   string
	Content string
	validator.Validator
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	posts, err := app.posts.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(w, r)
	data.Posts = posts

	app.render(w, http.StatusOK, "home.html", data)
}

func (app *application) postView(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	var id int
	var err error

	if idParam != "" {
		// Если параметр есть, преобразуем его в число
		id, err = strconv.Atoi(idParam)
	} else {
		// Иначе пытаемся извлечь ID из пути
		path := strings.TrimPrefix(r.URL.Path, "/post/view/")
		id, err = strconv.Atoi(path)
	}

	post, err := app.posts.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(w, r)
	data.Post = post

	app.render(w, http.StatusOK, "view.html", data)
}

func (app *application) postCreateForm(w http.ResponseWriter, r *http.Request) {
	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.flash(w, r, "To create post you need to login first.")
		// Если сессия отсутствует или недействительна, перенаправляем на логин
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return // Завершаем выполнение функции
	}

	// Используйте userID далее, если необходимо
	_ = userID // Уберите эту строку, если используете userID

	// Проверяем метод запроса
	if r.Method == http.MethodGet {
		data := app.newTemplateData(w, r)
		data.Form = &postCreateForm{
			Validator: validator.Validator{
				FieldErrors: map[string]string{},
			},
		}
		app.render(w, http.StatusOK, "create.html", data)
		return
	}

	// Если метод POST, обрабатываем данные формы
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		// Извлекаем данные из формы
		form := postCreateForm{
			Title:   r.PostForm.Get("title"),
			Content: r.PostForm.Get("content"),
		}

		// Валидация полей
		form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
		form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be longer than 100 characters")
		form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
		if !form.Valid() {
			data := app.newTemplateData(w, r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "create.html", data)
			return
		}

		// Вставляем данные в базу
		id, err := app.posts.Insert(form.Title, form.Content)
		if err != nil {
			app.serverError(w, err)
			return
		}

		app.flash(w, r, "Post created successfully!")
		// Перенаправляем пользователя на страницу с созданным постом
		http.Redirect(w, r, fmt.Sprintf("/post/view/%d", id), http.StatusSeeOther)
		return
	}

	// Если метод не GET и не POST, возвращаем ошибку
	app.clientError(w, http.StatusMethodNotAllowed)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	if r.Method == http.MethodGet {
		data := app.newTemplateData(w, r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	if r.Method == http.MethodPost {
		// Парсим данные из формы
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		form := userSignupForm{
			Name:      r.PostForm.Get("name"),
			Email:     r.PostForm.Get("email"),
			Password:  r.PostForm.Get("password"),
			Validator: validator.Validator{},
		}
		// Заполняем форму

		// Проверка на пустые поля и валидность
		form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
		form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
		form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
		form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
		form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

		// Если есть ошибки, возвращаем форму с ошибками
		if !form.Valid() {
			data := app.newTemplateData(w, r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
			return
		}

		// Вставка пользователя в базу данных
		err = app.users.Insert(form.Name, form.Email, form.Password)
		if err != nil {
			// Если ошибка вставки (например, email уже существует)
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(w, r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
			return
		}

		// Если регистрация прошла успешно, отображаем сообщение и редиректим
		app.flash(w, r, "Account created successfully!")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Для других методов (например, если GET запрос или некорректный метод)
	app.clientError(w, http.StatusMethodNotAllowed)
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data := app.newTemplateData(w, r)
		data.Form = userLoginForm{}
		app.render(w, http.StatusOK, "login.html", data)
	}
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		form := userSignupForm{
			Name:      r.PostForm.Get("name"),
			Email:     r.PostForm.Get("email"),
			Password:  r.PostForm.Get("password"),
			Validator: validator.Validator{},
		}

		form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
		form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
		form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
		if !form.Valid() {
			data := app.newTemplateData(w, r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.html", data)
			return
		}

		id, err := app.users.Authenticate(form.Email, form.Password)
		if err != nil {
			if errors.Is(err, models.ErrInvalidCredentials) {
				form.AddNonFieldError("Email or password is incorrect")
				data := app.newTemplateData(w, r)
				data.Form = form
				app.render(w, http.StatusUnprocessableEntity, "login.html", data)
				return
			} else {
				app.serverError(w, err)
				return
			}
		}

		app.flash(w, r, "Account logged in successfully!")
		app.setSession(w, id)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (app *application) userLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Renew the session token (invalidate the old session and create a new one).
	err := app.renewSessionToken(w, r)
	if err != nil {
		http.Error(w, "Failed to renew session token", http.StatusUnauthorized)
		return
	}

	// Remove the authenticated user ID from the session data.
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "No session found", http.StatusUnauthorized)
		return
	}

	// Remove the session from the session store.
	delete(app.sessions, cookie.Value)

	// Expire the session cookie on the client side.
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	// Add a flash message for the user.
	app.flash(w, r, "You've been logged out successfully!")

	// Redirect to the home page.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
