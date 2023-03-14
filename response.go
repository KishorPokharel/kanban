package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func (app *application) errorResponse(w http.ResponseWriter, status int, message string, err error) {
	app.logger.Println(err)
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	b := &bytes.Buffer{}
	out := map[string]any{
		"success": false,
		"message": message,
	}
	if err := json.NewEncoder(b).Encode(&out); err != nil {
		app.logger.Println(err)
		return
	}
	w.Write(b.Bytes())
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, out any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(&out); err != nil {
		app.errorResponse(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}
	w.Write(b.Bytes())
}
