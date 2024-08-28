package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LucasWiman90/GoWebApp/internal/config"
	"github.com/LucasWiman90/GoWebApp/internal/driver"
	"github.com/LucasWiman90/GoWebApp/internal/handlers"
	"github.com/LucasWiman90/GoWebApp/internal/helpers"
	"github.com/LucasWiman90/GoWebApp/internal/models"
	"github.com/LucasWiman90/GoWebApp/internal/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog, errorLog *log.Logger

// main is the main application function
func main() {
	db, err := run()

	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()

	defer close(app.MailChan)
	fmt.Println("Starting mail listener...")
	listenForMail()

	fmt.Printf("Starting application on port %s\n", portNumber)

	//Setup server
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB, error) {
	//Register models for sessions
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Reservation{})
	gob.Register(map[string]int{})
	//Change this to true when in production

	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	//Setup sessions
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	var password string
	fmt.Print("Enter password for PostgresSQL DB: ")
	fmt.Scanln(&password)

	connStr := fmt.Sprintf("host=172.25.32.1 port=5432 dbname=bookings user=postgres password=%s", password)

	//Connect to database
	log.Println("Connecting to database...")
	db, err := driver.ConnectSQL(connStr)
	if err != nil {
		log.Fatal("Cannot connect to database! Shutting down...")
	}
	log.Println("Connected to database!")

	//Create template cache
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatalf("cannot create template cache: %v", err)
		return nil, err
	}

	//Set the template cache as part of app config
	app.TemplateCache = tc
	app.UseCache = false

	//Create a repo for the appconfig and hand it back to the handlers
	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	//Sets the config for the rendering of templates
	render.NewRenderer(&app)

	//Sets the config for the usage of helpers
	helpers.NewHelpers(&app)

	return db, nil
}
