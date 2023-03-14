package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/KishorPokharel/kanban/postgres"
	validator "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func (app *application) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		app.errorResponse(
			w,
			http.StatusBadRequest,
			"Bad request body",
			fmt.Errorf("error: decoding json: %w", err),
		)
		return
	}
	if err := validator.ValidateStruct(&input,
		validator.Field(&input.Username, validator.Required, validator.Length(3, 20)),
		validator.Field(&input.Email, validator.Required, is.Email),
		validator.Field(&input.Password, validator.Required, validator.Length(10, 30)),
	); err != nil {
		out := map[string]any{
			"success": false,
			"errors":  err,
		}
		app.jsonResponse(w, http.StatusBadRequest, out)
		return
	}
	user := &postgres.User{
		Username: input.Username,
		Email:    input.Email,
	}
	err := user.Password.Set(input.Password)
	if err != nil {
		app.errorResponse(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	err = app.service.User.Create(user)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrDuplicateEmail):
			out := map[string]any{
				"success": false,
				"errors": map[string]any{
					"email": "email already exists",
				},
			}
			app.jsonResponse(w, http.StatusBadRequest, out)
			return
		default:
			app.errorResponse(w, http.StatusInternalServerError, "Something went wrong", err)
			return
		}
	}
	out := map[string]any{
		"success": true,
		"message": "User registered successfully",
		"data": map[string]any{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		},
	}
	app.jsonResponse(w, http.StatusCreated, out)
}

func (app *application) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		app.errorResponse(
			w,
			http.StatusBadRequest,
			"Bad request body",
			fmt.Errorf("error: decoding json: %w", err),
		)
		return
	}
	if err := validator.ValidateStruct(&input,
		validator.Field(&input.Email, validator.Required, is.Email),
		validator.Field(&input.Password, validator.Required),
	); err != nil {
		out := map[string]any{
			"success": false,
			"errors":  err,
		}
		app.jsonResponse(w, http.StatusBadRequest, out)
		return
	}
	user, err := app.service.User.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, postgres.ErrUserNotFound):
			out := map[string]any{
				"success": false,
				"message": "Invalid Credentials",
			}
			app.jsonResponse(w, http.StatusUnauthorized, out)
			return
		}
	}
	matches, err := user.Password.Matches(input.Password)
	if err != nil {
		app.errorResponse(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	if !matches {
		out := map[string]any{
			"success": false,
			"message": "Invalid Credentials",
		}
		app.jsonResponse(w, http.StatusUnauthorized, out)
		return
	}
	token, err := app.service.Token.New(user.ID, 24*time.Hour)
	if err != nil {
		app.errorResponse(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	out := map[string]any{
		"success": true,
		"data": map[string]any{
			"token": token.PlainText,
			"user": map[string]any{
				"id":         user.ID,
				"username":   user.Username,
				"email":      user.Email,
				"created_at": user.CreatedAt,
			},
		},
	}
	app.jsonResponse(w, http.StatusOK, out)
}
