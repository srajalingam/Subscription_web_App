package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// writeJSON is a helper method for writing JSON response
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(out)
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576 // 1MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

// badRequestResponse is a helper method for sending JSON response when the request is bad
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) error {
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = true
	payload.Message = err.Error()

	out, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(out)
	return nil
}

// invalidCredentialsResponse is a helper method for sending JSON response when the credentials are invalid
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) error {
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	payload.Error = true
	payload.Message = "Invalid credentials"

	err := app.writeJSON(w, http.StatusUnauthorized, payload)
	if err != nil {
		app.errorLog.Println(err)
		http.Error(w, "Unable to marshal json", http.StatusInternalServerError)
		return err
	}
	return nil
}
