package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	smtp struct {
		host     string
		port     int
		username string
		password string
	}
	frontendUrl string
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
}

func main() {

	////////////////////////////////////////////////////////////////
	// COMMAND LINE ARGUMENTS
	////////////////////////////////////////////////////////////////
	var cfg config
	flag.IntVar(&cfg.port, "port", 5000, "Server port to listen on")
	flag.Parse()

	/////////////////////////////////////////////////////////////////////////////
	// SETUP LOG FILES
	/////////////////////////////////////////////////////////////////////////////
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

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

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
	}

	//////////////////////////////////////////////////
	// CREATE PDF OUTPUT DIRECTORY IF IT DOESN'T EXIST
	//////////////////////////////////////////////////
	app.CreateDirIfNotExists("./invoices")

	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}

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

	app.infoLog.Printf("Starting invoice microservice on port %d", app.config.port)

	return srv.ListenAndServe()
}
