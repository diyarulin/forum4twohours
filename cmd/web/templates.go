package main

import "forum/internal/models"

// Структура хранилище для динамичных данных (модель поста..) которые мы хотим передать в хтмл шаблоны
type templateData struct {
	Post *models.Post
}
