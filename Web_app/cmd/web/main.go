package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
	"web_app/internal/driver"
	"web_app/internal/models"

	"github.com/alexedwards/scs/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

const version = "1.0.0"
const cssVersion = "1.0.0"

var session *scs.SessionManager

type config struct {
	port int
	env  string
	api  string
	db   struct {
		dsn string
	}
	stripe struct {
		secretKey string
		key       string
	}
}

type application struct {
	config        config
	infoLog       *log.Logger
	errorLog      *log.Logger
	templateCache map[string]*template.Template
	version       string
	DB            models.DBModel
	Session       *scs.SessionManager
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
	app.infoLog.Printf("Starting %s server on port %d in %s mode", version, app.config.port, app.config.env)
	return srv.ListenAndServe()
}

func main() {
	gob.Register(TransactionData{})
	// Initialize a new instance of the config struct.
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "dsn", os.Getenv("DATABASE_DSN"), "MySQL DSN")
	flag.StringVar(&cfg.api, "api", "http://localhost:4001", "API server URL")
	log.Println(os.Getenv("DATABASE_DSN"))
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	cfg.stripe.key = os.Getenv("STRIPE_PUBLISHABLE_KEY")

	// cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secretKey = os.Getenv("STRIPE_SECRET_KEY")

	cfg.api = os.Getenv("API")

	cfg.db.dsn = os.Getenv("DATABASE_DSN")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)

	conn, err := driver.OpenDB(cfg.db.dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer conn.Close()

	// Create a new session manager and configure it.
	session = scs.New()
	session.Lifetime = 24 * time.Hour

	tc := make(map[string]*template.Template)

	app := &application{
		config:        cfg,
		infoLog:       infoLog,
		errorLog:      errorLog,
		templateCache: tc,
		version:       version,
		DB:            models.DBModel{DB: conn},
		Session:       session,
	}

	err = app.serve()
	if err != nil {
		app.errorLog.Println(err)
		log.Fatal(err)
	}
}
