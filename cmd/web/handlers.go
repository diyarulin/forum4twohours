package main

import (
	"errors"
	"fmt"
	"forum/internal/models"
	"forum/internal/validator"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type postCreateForm struct {
	Title     string
	Content   string
	ImagePath string
	Author    string
	validator.Validator
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}
type userPost struct {
	ID        int
	Title     string
	Content   string
	ImagePath string
	Category  string
	Author    string
	Created   time.Time
}
type passwordForm struct {
	CurrentPassword     string `form:"currentPassword"`
	NewPassword         string `form:"newPassword"`
	ConfirmPassword     string `form:"confirmPassword"`
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
		err := r.ParseMultipartForm(20 << 20)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		file, handler, err := r.FormFile("image")
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}
		defer file.Close()
		app.infoLog.Printf("Uploaded File: %+v\n", handler.Filename)
		app.infoLog.Printf("File Size: %+v\n", handler.Size)
		app.infoLog.Printf("MIME Header: %+v\n", handler.Header)
		fileName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
		filePath := fmt.Sprintf("ui/static/upload/%s", fileName)
		if err := os.MkdirAll("ui/static/upload", os.ModePerm); err != nil {
			app.serverError(w, err)
			return
		}
		dst, err := os.Create(filePath)
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			app.serverError(w, err)
			return
		}
		id, err := app.getCurrentUser(r)
		if err != nil {
			app.serverError(w, err)
			return
		}
		author, err := app.users.Get(id)
		if err != nil {
			app.serverError(w, err)
			return
		}

		// Извлекаем данные из формы
		form := postCreateForm{
			Title:     r.PostForm.Get("title"),
			Content:   r.PostForm.Get("content"),
			ImagePath: filePath,
			Author:    author.Name,
		}
		form.ImagePath = strings.TrimPrefix(form.ImagePath, "ui/static/upload/")
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
		id, err = app.posts.Insert(form.Title, form.Content, form.ImagePath, form.Author)
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
func (app *application) profile(w http.ResponseWriter, r *http.Request) {

	id, err := app.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	user, err := app.users.Get(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	userPosts, err := app.posts.UserPosts(user.Name)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(w, r)
	data.Posts = userPosts
	data.User = &models.User{
		Name:  user.Name,
		Email: user.Email,
	}
	app.render(w, http.StatusOK, "profile.html", data)
}
func (app *application) changePassword(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		app.methodNotAllowed(w)
		return
	}
	id, err := app.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	form := passwordForm{
		CurrentPassword: r.FormValue("currentPassword"),
		NewPassword:     r.FormValue("newPassword"),
		ConfirmPassword: r.FormValue("confirmPassword"),
		Validator:       validator.Validator{},
	}
	// fmt.Println(form.CurrentPassword, form.ConfirmPassword, form.NewPassword)
	form.CheckField(validator.NotBlank(form.CurrentPassword), "currentPassword", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.CurrentPassword, 8), "currentPassword", "This field must be at least 8 characters long")
	form.CheckField(validator.NotBlank(form.NewPassword), "newPassword", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.NewPassword, 8), "newPassword", "This field must be at least 8 characters long")
	form.CheckField(validator.NotBlank(form.ConfirmPassword), "confirmPassword", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.ConfirmPassword, 8), "confirmPassword", "This field must be at least 8 characters long")
	form.CheckField(validator.ComparePassword(form.NewPassword, form.ConfirmPassword), "confirmPassword", "This field must be the same as newPassword")

	if !form.Validator.Valid() {
		data := app.newTemplateData(w, r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "profile.html", data)
		return
	}
	user, err := app.users.Get(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(form.CurrentPassword))
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		app.serverError(w, err)
		return
	}
	err = app.users.UpdatePassword(string(hashedPassword), id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Перенаправление на страницу профиля
	http.Redirect(w, r, "/user/profile/", http.StatusSeeOther)
}

// func (app *application) SaveEditPost(w http.ResponseWriter, r *http.Request) {
// if r.Method != http.MethodPost {
// 	app.methodNotAllowed(w)
// 	return
// }
// id, err := app.getCurrentUser(r)
// if err != nil {
// 	http.Redirect(w, r, "/", http.StatusFound)
// 	return
// }
// author, err := app.users.GetAuthor(id)
// if err != nil {
// 	app.serverError(w, err)
// 	return
// }
// // Получаем данные из формы
// postID := r.FormValue("id")         // ID поста
// postName := r.FormValue("Name")     // Название поста
// body := r.FormValue("Body")         // Текст поста
// category := r.FormValue("Category") // Категория поста

// // Проверка на пустые данные
// if postName == "" || body == "" || category == "" {
// 	http.Error(w, "Информация неполная", http.StatusBadRequest)
// 	return
// }

// // Получаем текущий пост
// strID, err := strconv.Atoi(postID)
// if err != nil {
// 	app.serverError(w, err)
// 	return
// }
// post, err := app.posts.Get(strID)
// if err != nil {
// 	app.serverError(w, err)
// 	return
// }
// if post.Author != author {
// 	http.Error(w, "Вы не можете редактировать этот пост", http.StatusForbidden)
// 	return
// }

// err := app.posts.UpdatePost(postName, body, )

// // Перенаправляем на страницу профиля
// http.Redirect(w, r, "/profile", http.StatusSeeOther)
// }
