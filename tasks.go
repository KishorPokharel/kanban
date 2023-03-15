package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/KishorPokharel/kanban/postgres"
	validator "github.com/go-ozzo/ozzo-validation/v4"
)

func (app *application) handleTasksGet(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	tasks, err := app.service.Task.GetAll(user.ID)
	if err != nil {
		app.errorResponse(
			w,
			http.StatusInternalServerError,
			"Internal Server Error",
			err,
		)
		return
	}
	out := map[string]any{
		"success": true,
		"data": map[string]any{
			"tasks": tasks,
		},
	}
	app.jsonResponse(w, http.StatusOK, out)
}

func (app *application) handleTaskCreate(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Content string `json:"content"`
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
		validator.Field(&input.Content, validator.Required),
	); err != nil {
		out := map[string]any{
			"success": false,
			"errors":  err,
		}
		app.jsonResponse(w, http.StatusBadRequest, out)
		return
	}
	user := app.contextGetUser(r)
	task := &postgres.Task{
		UserID:  user.ID,
		Content: input.Content,
	}
	if err := app.service.Task.Insert(task); err != nil {
		app.errorResponse(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	out := map[string]any{
		"success": true,
		"message": "Task added successfully",
		"data": map[string]any{
			"id":         task.ID,
			"content":    task.Content,
			"created_at": task.CreatedAt,
		},
	}
	app.jsonResponse(w, http.StatusCreated, out)
}

type sortInput struct {
	TaskID              int64  `json:"task_id"`
	SourceCategory      string `json:"source_category"`
	SourceIndex         int64  `json:"source_index"`
	DestinationCategory string `json:"destination_category"`
	DestinationIndex    int64  `json:"destination_index"`
}

func (app *application) handleTaskSort(w http.ResponseWriter, r *http.Request) {
	input := sortInput{}
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
		validator.Field(&input.TaskID, validator.Min(0)),
		validator.Field(&input.SourceCategory, validator.Required, validator.In("TODO", "DONE", "IN PROGRESS", "TESTING")),
		validator.Field(&input.SourceIndex, validator.Min(0)),
		validator.Field(&input.DestinationCategory, validator.Required, validator.In("TODO", "DONE", "IN PROGRESS", "TESTING")),
		validator.Field(&input.DestinationIndex, validator.Min(0)),
	); err != nil {
		out := map[string]any{
			"success": false,
			"errors":  err,
		}
		app.jsonResponse(w, http.StatusBadRequest, out)
		return
	}
	user := app.contextGetUser(r)
	if input.SourceCategory == input.DestinationCategory {
		if err := app.service.Task.SortTaskInSameCategory(
			user.ID,
			input.TaskID,
			input.SourceIndex,
			input.DestinationIndex,
			input.DestinationCategory,
		); err != nil {
			switch {
			case errors.Is(err, postgres.ErrInvalidData):
				app.errorResponse(w, http.StatusBadRequest, "invalid data", err)
				return
			default:
				app.errorResponse(w, http.StatusInternalServerError, "Something went wrong", err)
				return
			}
		}
		out := map[string]any{
			"success": true,
		}
		app.jsonResponse(w, http.StatusOK, out)
		return
	} else {
		if err := app.service.Task.SortTaskInDifferentCategory(
			user.ID,
			input.TaskID,
			input.SourceIndex,
			input.DestinationIndex,
			input.SourceCategory,
			input.DestinationCategory,
		); err != nil {
			switch {
			case errors.Is(err, postgres.ErrInvalidData):
				app.errorResponse(w, http.StatusBadRequest, "invalid data", err)
				return
			default:
				app.errorResponse(w, http.StatusInternalServerError, "Something went wrong", err)
				return
			}
		}
		out := map[string]any{
			"success": true,
		}
		app.jsonResponse(w, http.StatusOK, out)
		return
	}
}
