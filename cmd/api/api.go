package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/leetrent/go-stripe/internal/driver"
	"github.com/leetrent/go-stripe/internal/models"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		key    string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
	}
	secretKey   string
	frontendUrl string
	invoiceUrl  string
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
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Printf("Starting API server in %s mode on port %d", app.config.env, app.config.port)

	return srv.ListenAndServe()
}
func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4001, "Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application environment {development|production|maintenance}")
	flag.StringVar(&cfg.db.dsn, "dsn", "not provided", "DB Connection String")

	// SMTP
	// flag.StringVar(&cfg.smtp.host, "smtphost", "smtp.mailtrap.io", "smtp host")
	// flag.IntVar(&cfg.smtp.port, "smtpport", 587, "smtp port")

	flag.Parse()

	fmt.Printf("\n[api][main] => (cfg.db.dsn): %s\n", cfg.db.dsn)

	/////////////////////////////////////////////////////////////////////////////
	// SETUP LOG FILES
	/////////////////////////////////////////////////////////////////////////////
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	/////////////////////////////////////////////////////////////////////////////
	// STRIPE ENVIRONMENT VARIABLES
	/////////////////////////////////////////////////////////////////////////////
	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")

	/////////////////////////////////////////////////////////////////////////////
	// SMTP ENVIRONMENT VARIABLES
	/////////////////////////////////////////////////////////////////////////////
	cfg.smtp.host = os.Getenv("SMTP_HOST")
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		errorLog.Fatal(err)
	}
	cfg.smtp.port = smtpPort
	cfg.smtp.username = os.Getenv("SMTP_USERNAME")
	cfg.smtp.password = os.Getenv("SMTP_PASSWORD")

	/////////////////////////////////////////////////////////////////////////////
	// CHANGE PASSWORD ENVIRONMENT VARIABLES
	/////////////////////////////////////////////////////////////////////////////
	cfg.secretKey = os.Getenv("SECRET_KEY")
	cfg.frontendUrl = os.Getenv("FRONTEND_URL")

	/////////////////////////////////////////////////////////////////////////////
	// INVOICE MICROSERVICE URL
	/////////////////////////////////////////////////////////////////////////////
	cfg.invoiceUrl = os.Getenv("INVOICE_URL")

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
		log.Fatal(err)
	}
}
