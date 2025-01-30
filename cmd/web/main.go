package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"forum/internal/models"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	errorLog           *log.Logger
	infoLog            *log.Logger
	posts              *models.PostModel
	users              *models.UserModel
	comments           *models.CommentModel
	categories         *models.CategoryModel
	reactions          *models.ReactionModel
	notificationsModel *models.NotificationModel
	templateCache      map[string]*template.Template
	sessions           map[string]int
	mu                 sync.Mutex
	reports            *models.ReportModel
}

func main() {
	// Адрес порта
	addr := flag.String("addr", ":4000", "http service address")
	dsn := "./data/forum.db"
	flag.Parse()

	// Логгеры для ошибок и информации
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Открытие базы данных
	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	// Инициализация кэша шаблонов
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// Инициализация структуры приложения
	app := application{
		errorLog:           errorLog,
		infoLog:            infoLog,
		posts:              &models.PostModel{DB: db},
		users:              &models.UserModel{DB: db},
		comments:           &models.CommentModel{DB: db},
		categories:         &models.CategoryModel{DB: db},
		notificationsModel: &models.NotificationModel{DB: db},
		reactions:          &models.ReactionModel{DB: db},
		templateCache:      templateCache,
		sessions:           make(map[string]int),
		reports:            &models.ReportModel{DB: db}, // Добавляем поле reports корректно
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Инициализация структуры сервера для использования errorLog и роутера
	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Запуск сервера с поддержкой HTTPS
	infoLog.Printf("Starting server on https://localhost%s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
