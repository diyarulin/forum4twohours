package main

import (
	"fmt"
	"forum/internal/models"
	"html/template"
	"path/filepath"
	"time"
)

// templateData — структура для хранения данных, передаваемых в HTML-шаблоны
type templateData struct {
	CurrentYear         int
	Post                *models.Post   // Один пост (для страницы просмотра одного поста)
	Posts               []*models.Post // Список постов (например, для главной страницы)
	User                *models.User
	Users               []*models.User
	Comment             *models.Comment
	Comments            []*models.Comment
	Notifications       []*models.Notification
	UnreadNotifications int
	Categories          []*models.Category
	Form                any
	IsLiked             bool
	IsDisliked          bool
	SelectedCategory    string
	Flash               string
	IsAuthenticated     bool
	Status              int
	Message             string
	PendingPosts        []*models.Post
	Reports             []*models.Report
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

// newTemplateCache создаёт кэш шаблонов, чтобы не парсить их каждый раз
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// Специальная обработка для страницы ошибок
		if name == "errors.html" {
			ts, err := template.New(name).Funcs(functions).ParseFiles(page)
			if err != nil {
				return nil, err
			}
			cache[name] = ts
			continue
		}

		// Обычные страницы
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	// Добавляем шаблон ошибок, если он не был найден
	if _, ok := cache["errors.html"]; !ok {
		ts, err := template.New("errors.html").ParseFiles("./ui/html/pages/errors.html")
		if err != nil {
			return nil, fmt.Errorf("error template not found: %v", err)
		}
		cache["errors.html"] = ts
	}

	return cache, nil
}

// newTemplateCache собирает и кэширует все шаблоны страниц вместе с базовым шаблоном и навигацией.
// Это повышает производительность приложения, избегая повторного парсинга шаблонов при каждом запросе.
