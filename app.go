package main

import (
	"log"
	"net/http"

	"github.com/KishorPokharel/kanban/postgres"
	"github.com/julienschmidt/httprouter"
)

type application struct {
	logger  *log.Logger
	service postgres.Service
}

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodPost, "/api/users/register", app.handleUserRegister)
	router.HandlerFunc(http.MethodPost, "/api/users/login", app.handleUserLogin)

	router.HandlerFunc(http.MethodGet, "/api/tasks", app.authenticate(app.handleTasksGet))
	router.HandlerFunc(http.MethodPost, "/api/tasks", app.authenticate(app.handleTaskCreate))
	router.HandlerFunc(http.MethodPost, "/api/tasks/sort", app.authenticate(app.handleTaskSort))

	return app.logRequest(app.enableCors(router))
}

func (app *application) run() error {
	app.logger.Println("app running")
	return http.ListenAndServe(":3000", app.routes())
}
