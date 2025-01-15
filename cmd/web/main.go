package main

import (
	"database/sql"
	"flag"
	"forum/internal/models"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"os"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	posts         *models.PostModel
	templateCache map[string]*template.Template
}

func main() {
	// Адрес порта
	addr := flag.String("addr", ":4000", "http service address")
	dsn := "./forum.db"

	flag.Parse()

	// Логгеры для ошибок и информации
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

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

	// Инициализация структуры приложения для того что бы хэндлеры применялись как методы к этой структуре и видели еррорлог и инфолог
	app := application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		posts:         &models.PostModel{DB: db},
		templateCache: templateCache,
	}

	// Инициализация структуры сервера для использования errorLog и роутера
	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on http://localhost%s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
