package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LucasWiman90/GoWebApp/internal/config"
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
	err := run()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting application on port %s\n", portNumber)

	//Setup server
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() error {
	gob.Register(models.Reservation{})
	//Change this to true when in production
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

	//Create template cache
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
		return err
	}

	//Set the template cache as part of app config
	app.TemplateCache = tc
	app.UseCache = false

	//Create a repo for the appconfig and hand it back to the handlers
	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	//Sets the config for the rendering of templates
	render.NewTemplates(&app)

	//Sets the config for the usage of helpers
	helpers.NewHelpers(&app)

	return nil
}
