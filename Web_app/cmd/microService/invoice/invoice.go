package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

type Config struct {
	port int
	smtp struct {
		host     string
		port     int
		username string
		password string
	}
	frontend string
}

type application struct {
	config   Config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
}

func main() {
	var cfg Config

	flag.IntVar(&cfg.port, "port", 5000, "API server port")
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "your-smtp-username", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "your-smtp-password", "SMTP password")
	flag.StringVar(&cfg.frontend, "fr	ontend", "http://localhost:3000", "Frontend URL")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
	}
	err := app.serve()
	if err != nil {
		log.Fatal(err)
	}
}

// serve
func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	app.infoLog.Printf("Starting Invoice API server on port %d", app.config.port)
	return srv.ListenAndServe()
}
