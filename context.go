package main

import (
	"context"
	"net/http"

	"github.com/KishorPokharel/kanban/postgres"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *postgres.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *postgres.User {
	user, ok := r.Context().Value(userContextKey).(*postgres.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
