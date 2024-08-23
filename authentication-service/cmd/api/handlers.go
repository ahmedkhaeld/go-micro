package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (app *App) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	//read the request data into payload
	err := app.ReadJSON(w, r, &requestPayload, true)
	if err != nil {
		app.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	//validate the user credentials
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		app.ErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	isValid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !isValid {
		app.ErrorJSON(w, errors.New("invalid credentials"), http.StatusInternalServerError)
		return
	}

	//log the authentication data to the logger service
	// log authentication
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		app.ErrorJSON(w, errors.New("error log the data to mongodb: "+err.Error()), http.StatusInternalServerError)
		return
	}

	payload := response{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.WriteJSON(w, http.StatusAccepted, payload)

}

func (app *App) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service:8080/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
