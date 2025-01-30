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
	Category  string
	Author    string
	AuthorID  int
	validator.Validator
	Status string
}
type editPost struct {
	ID        int
	Title     string
	Content   string
	ImagePath string
	Category  string
	Author    string
	AuthorID  int
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
type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) manageUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.methodNotAllowed(w)
		return
	}
	if r.URL.Path != "/admin/users" {
		http.NotFound(w, r)
		return
	}
	users, err := app.users.GetPendingModerators()
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(w, r)

	data.Users = users
	app.render(w, http.StatusOK, "users.html", data)
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.methodNotAllowed(w)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	selectedCategory := r.URL.Query().Get("Category")
	data := app.newTemplateData(w, r)
	var posts []*models.Post
	var err error
	if selectedCategory == "" {
		posts, err = app.posts.Latest()
	} else {
		posts, err = app.posts.SortByCategory(selectedCategory)
		data.SelectedCategory = selectedCategory
	}
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.Posts = posts
	categories, err := app.categories.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}
	userID, err := app.getCurrentUser(r)
	if err != nil && userID == 0 {
		data.Categories = categories
		data.IsAuthenticated = app.isAuthenticated(r)
		app.render(w, http.StatusOK, "home.html", data)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}
	user, err := app.users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data.Categories = categories
	data.User = user
	data.IsAuthenticated = app.isAuthenticated(r)
	app.render(w, http.StatusOK, "home.html", data)
	return
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

	comments, err := app.comments.GetByPostID(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(w, r)
	data.Post = post
	data.Comments = comments
	if app.isAuthenticated(r) {
		userId, err := app.getCurrentUser(r)
		if err != nil {
			app.clientError(w, http.StatusUnauthorized)
		}
		user, err := app.users.Get(userId)
		if err != nil {
			app.serverError(w, err)
		}
		data.User = user
	}
	data.IsAuthenticated = app.isAuthenticated(r)
	app.render(w, http.StatusOK, "view.html", data)
}

func (app *application) postCreateForm(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		categories, err := app.categories.GetAll()
		if err != nil {
			app.serverError(w, err)
			return
		}

		data := app.newTemplateData(w, r)
		data.Form = &postCreateForm{
			Validator: validator.Validator{
				FieldErrors: map[string]string{},
			},
		}
		data.Categories = categories
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

		// Пытаемся получить файл, но проверяем, прикреплен ли он
		file, handler, err := r.FormFile("image")
		if err != nil && err.Error() != "http: no such file" { // проверяем, что файл не был прикреплен
			app.clientError(w, http.StatusBadRequest)
			return
		}

		var filePath string
		if err == nil {
			// Файл был прикреплен, обрабатываем его
			defer file.Close()
			app.infoLog.Printf("Uploaded File: %+v\n", handler.Filename)
			app.infoLog.Printf("File Size: %+v\n", handler.Size)
			app.infoLog.Printf("MIME Header: %+v\n", handler.Header)
			fileName := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
			filePath = fmt.Sprintf("ui/static/upload/%s", fileName)
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
		} else {
			// Если файл не был прикреплен, оставляем filePath пустым
			filePath = ""
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
		var statusString string
		if author.Role == "moderator" || author.Role == "admin" {
			statusString = "approved"
		} else {
			statusString = "pending"
		}
		form := postCreateForm{
			Title:     r.PostForm.Get("title"),
			Content:   r.PostForm.Get("content"),
			ImagePath: filePath,
			Category:  r.PostForm.Get("Category"),
			Author:    author.Name,
			AuthorID:  id,
			Status:    statusString,
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

		app.infoLog.Printf("User Role: %s, Setting post status to: %s", author.Role, form.Status)
		// Вставляем данные в базу
		id, err = app.posts.Insert(
			form.Title,
			form.Content,
			form.ImagePath,
			form.Category,
			form.Author,
			form.Status,
			form.AuthorID,
		)

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
		form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address xxx@xxx.xx")
		form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
		form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
		form.CheckField(validator.ValidatePassword(form.Password), "password", "Password must contain at least one lowercase letter, one uppercase letter, one digit, and one special character (e.g., @, #, $, %).")

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

	userPosts, err := app.posts.UserPosts(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	userComments, err := app.comments.UserComments(id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(w, r)
	data.Posts = userPosts
	data.Comments = userComments
	data.User = &models.User{
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
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
	app.flash(w, r, "Password changed successfully!")
	// Перенаправление на страницу профиля
	http.Redirect(w, r, "/user/profile/", http.StatusSeeOther)
}

func (app *application) EditPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		app.methodNotAllowed(w)
		return
	}
	app.infoLog.Printf("Запрос: Метод: %s, Путь: %s", r.Method, r.URL.Path)
	if r.Method == http.MethodGet {
		idParam := r.URL.Query().Get("id")
		var id int
		var err error

		if idParam != "" {
			// Если параметр есть, преобразуем его в число
			id, err = strconv.Atoi(idParam)
			if err != nil {
				app.serverError(w, err)
			}
		} else {
			// Иначе пытаемся извлечь ID из пути
			path := strings.TrimPrefix(r.URL.Path, "/post/edit/")
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
		data.Form = editPost{
			ID:        post.ID,
			Title:     post.Title,
			Content:   post.Content,
			ImagePath: post.ImagePath,
			Category:  post.Category,
			Author:    post.Author,
			AuthorID:  post.AuthorID,
		}

		app.render(w, http.StatusOK, "edit_post.html", data)
	}

	// Если метод POST, обрабатываем данные формы
	if r.Method == http.MethodPost {
		err := r.ParseMultipartForm(20 << 20)
		if err != nil {
			app.clientError(w, http.StatusBadRequest)
			return
		}

		var filePath string
		var fileName string
		file, handler, err := r.FormFile("image")

		if err != nil {
			if err != http.ErrMissingFile { // Если ошибка не связана с отсутствием файла
				app.clientError(w, http.StatusBadRequest)
				return
			}
		} else {
			// Если файл загружен
			defer file.Close()
			app.infoLog.Printf("Uploaded File: %+v\n", handler.Filename)
			app.infoLog.Printf("File Size: %+v\n", handler.Size)
			app.infoLog.Printf("MIME Header: %+v\n", handler.Header)
			fileName = fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
			filePath = fmt.Sprintf("ui/static/upload/%s", fileName)
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
		}

		id, err := app.getCurrentUser(r)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		author, err := app.users.Get(id)
		if err != nil {
			app.serverError(w, err)
			return
		}
		strID := r.PostForm.Get("id")
		intID, err := strconv.Atoi(strID)
		if err != nil {
			app.serverError(w, err)
			return
		}
		form := editPost{
			ID:        intID,
			Title:     r.PostForm.Get("title"),
			Content:   r.PostForm.Get("content"),
			ImagePath: filePath, // Путь к изображению только если файл был загружен
			Category:  r.PostForm.Get("category"),
			Author:    author.Name,
			AuthorID:  author.ID,
		}

		form.ImagePath = strings.TrimPrefix(form.ImagePath, "ui/static/upload/")
		// Валидация полей
		form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
		form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be longer than 100 characters")
		form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
		if !form.Valid() {
			data := app.newTemplateData(w, r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "edit_post.html", data)
			return
		}

		// Получаем текущий пост

		post, err := app.posts.Get(form.ID)
		if err != nil {
			app.serverError(w, err)
			return
		}
		if form.ImagePath == "" {
			form.ImagePath = post.ImagePath
		}
		app.infoLog.Printf("Updating post: title=%s, content=%s, imagePath=%s, category=%s, author=%s", form.Title, form.Content, form.ImagePath, form.Category, form.Author)
		err = app.posts.UpdatePost(form.Title, form.Content, form.ImagePath, form.Category, form.Author, form.AuthorID, form.ID)
		if err != nil {
			app.serverError(w, err)
		}
		app.flash(w, r, "Post edited successfully!")
		// Перенаправляем на страницу профиля
		http.Redirect(w, r, fmt.Sprintf("/post/view/%d", form.ID), http.StatusSeeOther)
		return
	}

}
func (app *application) DeletePost(w http.ResponseWriter, r *http.Request) {
	// Получаем ID поста из URL
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/post/delete/"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Получаем текущего пользователя
	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	// Проверяем права доступа
	post, err := app.posts.Get(id)
	if err != nil {
		app.notFound(w)
		return
	}

	user, err := app.users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Разрешаем удаление если:
	// 1. Пользователь - автор поста
	// 2. Пользователь модератор или админ
	if post.AuthorID != userID && user.Role != "moderator" && user.Role != "admin" {
		app.clientError(w, http.StatusForbidden)
		return
	}

	// Логика удаления поста
	path, err := app.posts.DeletePost(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Удаление файла если есть
	if path != "" {
		if err := os.Remove("./ui/static/upload/" + path); err != nil {
			app.errorLog.Println("Failed to delete image:", err)
		}
	}

	app.flash(w, r, "Post deleted successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (app *application) getComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		app.methodNotAllowed(w)
		return
	}

	postIDStr := r.URL.Query().Get("post_id")
	if postIDStr == "" {
		http.Error(w, "post_id is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "invalid post_id", http.StatusBadRequest)
		return
	}

	comments, err := app.comments.GetByPostID(postID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(w, r)
	data.Comments = comments
	app.render(w, http.StatusOK, "comments.html", data)
}
func (app *application) addComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.methodNotAllowed(w)
		return
	}
	idParam := r.FormValue("post_id")
	var id int
	var err error

	if idParam != "" {
		id, err = strconv.Atoi(idParam)
	} else {
		path := strings.TrimPrefix(r.URL.Path, "/post/view/")
		id, err = strconv.Atoi(path)
	}
	content := r.FormValue("content")
	// author := r.FormValue("author")
	user_id, err := app.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	user, err := app.users.Get(user_id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	comment := &models.Comment{
		PostID:  id,
		UserID:  user_id,
		Author:  user.Name,
		Content: content,
		Created: time.Now(),
	}

	// Сохраняем комментарий в базе данных
	err = app.comments.Insert(comment)
	if err != nil {
		app.serverError(w, err)
		return
	}
	post, err := app.posts.Get(comment.PostID)
	if err == nil && post.AuthorID != comment.UserID {
		// Создаем уведомление для автора поста
		err = app.notificationsModel.Insert(
			post.AuthorID,
			comment.UserID,
			"comment",
			post.ID,
			comment.ID,
		)
		if err != nil {
			app.errorLog.Println("Failed to create notification:", err)
		}
	}
	// Перенаправляем на страницу поста с комментариями
	app.flash(w, r, "Comment added successfully!")
	http.Redirect(w, r, fmt.Sprintf("/post/view/%d", id), http.StatusSeeOther)
	return
}

func (app *application) deleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.methodNotAllowed(w)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	commentIDStr := r.Form.Get("comment_id")
	postIDStr := r.Form.Get("post_id")

	if commentIDStr == "" || postIDStr == "" {
		app.errorLog.Print("DELETE COMMENT: comment_id and post_id are required")
		app.errorLog.Println(w, http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		app.errorLog.Print("DELETE COMMENT: invalid comment_id")
		app.errorLog.Println(w, http.StatusBadRequest)
		return
	}

	// Удаляем комментарий из базы
	err = app.comments.Delete(commentID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Перенаправляем на страницу поста
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		app.errorLog.Print("DELETE COMMENT: invalid post_id")
		app.errorLog.Println(w, http.StatusBadRequest)
		return
	}
	app.flash(w, r, "Comment deleted successfully!")
	http.Redirect(w, r, fmt.Sprintf("/post/view/%d", postID), http.StatusSeeOther)
}
func (app *application) likePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.methodNotAllowed(w)
		return
	}

	postIDStr := r.FormValue("post_id")
	if postIDStr == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err = app.reactions.LikePost(postID, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	post, err := app.posts.Get(postID)
	if err == nil && post.AuthorID != userID {
		err = app.notificationsModel.Insert(
			post.AuthorID,
			userID,
			"post_like",
			postID,
			0,
		)
		if err != nil {
			app.errorLog.Println("Failed to create notification:", err)
		}
	}

	app.flash(w, r, "Post liked successfully!")
	http.Redirect(w, r, fmt.Sprintf("/post/view/%d", postID), http.StatusSeeOther)
}

func (app *application) dislikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.methodNotAllowed(w)
		return
	}

	postIDStr := r.FormValue("post_id")
	if postIDStr == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err = app.reactions.DislikePost(postID, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	post, err := app.posts.Get(postID)
	if err == nil && post.AuthorID != userID {
		err = app.notificationsModel.Insert(
			post.AuthorID,
			userID,
			"post_dislike",
			postID,
			0,
		)
		if err != nil {
			app.errorLog.Println("Failed to create notification:", err)
		}
	}
	app.flash(w, r, "Post disliked successfully!")
	http.Redirect(w, r, fmt.Sprintf("/post/view/%d", postID), http.StatusSeeOther)
}

func (app *application) removeLikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.methodNotAllowed(w)
		return
	}

	postIDStr := r.FormValue("post_id")
	if postIDStr == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err = app.reactions.RemoveLikePost(postID, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.flash(w, r, "Like removed successfully!")
	http.Redirect(w, r, fmt.Sprintf("/post/view/%d", postID), http.StatusSeeOther)
}

func (app *application) removeDislikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.methodNotAllowed(w)
		return
	}

	postIDStr := r.FormValue("post_id")
	if postIDStr == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err = app.reactions.RemoveDislikePost(postID, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.flash(w, r, "Dislike removed successfully!")
	http.Redirect(w, r, fmt.Sprintf("/post/view/%d", postID), http.StatusSeeOther)
}
func (app *application) likeComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.methodNotAllowed(w)
		return
	}

	commentIDStr := r.FormValue("comment_id")
	if commentIDStr == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err = app.reactions.LikeComment(commentID, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	comment, err := app.comments.GetByID(commentID)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	post, err := app.posts.Get(comment.PostID)
	if err == nil && post.AuthorID != comment.UserID {
		// Создаем уведомление для автора поста
		err = app.notificationsModel.Insert(
			post.AuthorID,
			comment.UserID,
			"comment_like",
			post.ID,
			comment.ID,
		)
		if err != nil {
			app.errorLog.Println("Failed to create notification:", err)
		}
	}
	app.flash(w, r, "Comment liked successfully!")
	http.Redirect(w, r, fmt.Sprintf("/post/view/%d", comment.PostID), http.StatusSeeOther)
}

func (app *application) dislikeComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.methodNotAllowed(w)
		return
	}

	commentIDStr := r.FormValue("comment_id")
	if commentIDStr == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err = app.reactions.DislikeComment(commentID, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	comment, err := app.comments.GetByID(commentID)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	post, err := app.posts.Get(comment.PostID)
	if err == nil && post.AuthorID != comment.UserID {
		// Создаем уведомление для автора поста
		err = app.notificationsModel.Insert(
			post.AuthorID,
			comment.UserID,
			"comment_dislike",
			post.ID,
			comment.ID,
		)
		if err != nil {
			app.errorLog.Println("Failed to create notification:", err)
		}
	}
	app.flash(w, r, "Comment disliked successfully!")
	http.Redirect(w, r, fmt.Sprintf("/post/view/%d", comment.PostID), http.StatusSeeOther)
}

func (app *application) removeLikeComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.methodNotAllowed(w)
		return
	}

	commentIDStr := r.FormValue("comment_id")
	if commentIDStr == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err = app.reactions.RemoveLikeComment(commentID, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	comment, err := app.comments.GetByID(commentID)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	app.flash(w, r, "Like removed successfully!")
	http.Redirect(w, r, fmt.Sprintf("/post/view/%d", comment.PostID), http.StatusSeeOther)
}

func (app *application) removeDislikeComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		app.methodNotAllowed(w)
		return
	}

	commentIDStr := r.FormValue("comment_id")
	if commentIDStr == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	err = app.reactions.RemoveDislikeComment(commentID, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	comment, err := app.comments.GetByID(commentID)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	app.flash(w, r, "Dislike removed successfully!")
	http.Redirect(w, r, fmt.Sprintf("/post/view/%d", comment.PostID), http.StatusSeeOther)
}
func (app *application) notifications(w http.ResponseWriter, r *http.Request) {

	if !app.isAuthenticated(r) {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	userID, err := app.getCurrentUser(r)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Получаем уведомления
	notifications, err := app.notificationsModel.GetAll(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Помечаем как прочитанные
	err = app.notificationsModel.MarkAllAsRead(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(w, r)
	data.Notifications = notifications
	app.render(w, http.StatusOK, "notifications.html", data)
}
func (app *application) manageCategories(w http.ResponseWriter, r *http.Request) {
	// Проверка прав администратора
	userID, err := app.getCurrentUser(r)
	if err != nil || !app.isAdmin(userID) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	categories, err := app.categories.GetAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(w, r)
	data.Categories = categories
	app.render(w, http.StatusOK, "categories.html", data)
}

func (app *application) addCategory(w http.ResponseWriter, r *http.Request) {
	if !app.isAdminRequest(r) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	name := r.FormValue("name")
	if err := app.categories.Insert(name); err != nil {
		if errors.Is(err, models.ErrDuplicateCategory) {
			app.flash(w, r, "Category already exists!")
		} else {
			app.serverError(w, err)
		}
	}
	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func (app *application) updateCategory(w http.ResponseWriter, r *http.Request) {
	if !app.isAdminRequest(r) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	newName := r.FormValue("name")
	if err := app.categories.Update(id, newName); err != nil {
		app.serverError(w, err)
	}
	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}

func (app *application) deleteCategory(w http.ResponseWriter, r *http.Request) {
	if !app.isAdminRequest(r) {
		app.clientError(w, http.StatusForbidden)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	if err := app.categories.Delete(id); err != nil {
		app.serverError(w, err)
	}
	http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
}
