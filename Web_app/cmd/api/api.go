package main

import (
	_ "github.com/go-sql-driver/mysql"

	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"web_app/internal/driver"
	"web_app/internal/models"

	"github.com/joho/godotenv"
)

const version = "1.0.0"
const cssVersion = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	stripe struct {
		secretKey string
		key       string
	}
	smtp struct {
		host     string
		port     string
		username string
		password string
	}
	secretKey string
	frontend  string
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
	DB       models.DBModel
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		ErrorLog:          app.errorLog,
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
	app.infoLog.Printf("Starting  Backend %s server on port %d in %s mode", version, app.config.port, app.config.env)
	return srv.ListenAndServe()
}
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
		log.Println("No .env file found")
	}
	var cfg config
	flag.IntVar(&cfg.port, "port", 4001, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "dsn", os.Getenv("DATABASE_DSN"), "MySQL DSN")
	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	flag.StringVar(&cfg.smtp.port, "smtp-port", os.Getenv("SMTP_PORT"), "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")

	flag.StringVar(&cfg.secretKey, "secret", os.Getenv("SECRET_KEY"), "Secret key for session management")
	flag.StringVar(&cfg.frontend, "frontend", os.Getenv("FRONTEND_URL"), "Frontend URL")

	flag.Parse()

	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secretKey = os.Getenv("STRIPE_SECRET")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)

	conn, err := driver.OpenDB(cfg.db.dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer conn.Close()

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
		DB:       models.DBModel{DB: conn},
	}

	err = app.serve()
	if err != nil {
		app.errorLog.Fatal(err)
	}
}
