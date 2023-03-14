package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/KishorPokharel/kanban/postgres"
)

func (app *application) enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ah := r.Header.Get("Authorization")
		if ah == "" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			app.errorResponse(w,
				http.StatusUnauthorized,
				"Invalid or Missing Auth Token",
				errors.New("invalid or missing authorization token"),
			)
			return
		}
		fields := strings.Fields(ah)
		if len(fields) != 2 || fields[0] != "Bearer" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			app.errorResponse(w,
				http.StatusUnauthorized,
				"Invalid or Missing Auth Token",
				errors.New("missing authorization token"),
			)
			return
		}
		token := fields[1]
		user, err := app.service.User.GetForToken(token)
		if err != nil {
			switch {
			case errors.Is(err, postgres.ErrUserNotFound):
				w.Header().Set("WWW-Authenticate", "Bearer")
				app.errorResponse(w,
					http.StatusUnauthorized,
					"Invalid Token",
					errors.New("missing authorization token"),
				)
				return
			default:
				app.errorResponse(w, http.StatusInternalServerError, "Something went wrong", err)
				return
			}
		}
		r = app.contextSetUser(r, user)
		hf(w, r)
	}
}

func (app *application) logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		handler.ServeHTTP(w, r)
		elapsed := time.Since(now)
		app.logger.Printf("%s %s %s\n", r.Method, r.URL, elapsed)
	})
}
