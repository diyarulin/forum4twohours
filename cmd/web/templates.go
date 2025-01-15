package main

import (
	"forum/internal/models"
	"html/template"
	"path/filepath"
	"time"
)

// templateData — структура для хранения данных, передаваемых в HTML-шаблоны
type templateData struct {
	CurrentYear int
	Post        *models.Post   // Один пост (для страницы просмотра одного поста)
	Posts       []*models.Post // Список постов (например, для главной страницы)
}

func humanDate(t time.Time) string {
	return t.Format("11 Jan 22006 at 05:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

// newTemplateCache создаёт кэш шаблонов, чтобы не парсить их каждый раз
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// Find all the HTML files for pages
	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// Parse the base template
		ts, err := template.ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		// Parse partial templates and page-specific template
		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Store the template set in the cache
		cache[name] = ts
	}
	return cache, nil
}

// newTemplateCache собирает и кэширует все шаблоны страниц вместе с базовым шаблоном и навигацией.
// Это повышает производительность приложения, избегая повторного парсинга шаблонов при каждом запросе.