package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/LucasWiman90/GoWebApp/pkg/config"
	"github.com/LucasWiman90/GoWebApp/pkg/handlers"
	"github.com/LucasWiman90/GoWebApp/pkg/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager

// main is the main application function
func main() {
	//Change this to true when in production
	app.InProduction = false

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
	}

	//Set the template cache as part of app config
	app.TemplateCache = tc
	app.UseCache = false

	//Create a repo for the appconfig and hand it back to the handlers
	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	//Sets the config for the rendering of templates
	render.NewTemplates(&app)

	fmt.Printf("Starting application on port %s\n", portNumber)

	//Setup server
	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	log.Fatal(err)
}
