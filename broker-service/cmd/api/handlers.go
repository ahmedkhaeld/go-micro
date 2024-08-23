package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *App) Broker(w http.ResponseWriter, r *http.Request) {
	payload := responsePayload{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.WriteJSON(w, http.StatusOK, payload)
}

func (app *App) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.ReadJSON(w, r, &requestPayload)
	if err != nil {
		app.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.Authenticate(w, requestPayload.Auth)
	default:
		app.ErrorJSON(w, fmt.Errorf("unknown action: %s", requestPayload.Action), http.StatusBadRequest)
		return
	}

}

func (app *App) Authenticate(w http.ResponseWriter, a AuthPayload) {

	// create some json we'll send to the auth microservice
	jsonData, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		app.ErrorJSON(w, errors.New("could not marshal json"), http.StatusInternalServerError)
		return
	}

	//call the auth service
	url := "http://auth-service:8080/auth"

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJSON(w, errors.New("could not send post req"), http.StatusInternalServerError)
		return
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.ErrorJSON(w, errors.New("error with from remote response"), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	//make sure the response is correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.ErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.ErrorJSON(w, fmt.Errorf("unexpected status code: %d", response.StatusCode), http.StatusInternalServerError)
		return
	}

	//create var to read response.Body into
	var jsonFromRemote responsePayload

	//read the response body
	err = json.NewDecoder(response.Body).Decode(&jsonFromRemote)
	if err != nil {
		app.ErrorJSON(w, errors.New("error decoding remote response"), http.StatusInternalServerError)
		return
	}

	if jsonFromRemote.Error {
		app.ErrorJSON(w, errors.New(jsonFromRemote.Message), http.StatusUnauthorized)
		return
	}

	var payload responsePayload
	payload.Data = jsonFromRemote.Data
	payload.Error = false
	payload.Message = fmt.Sprintf("Authenticated user %s", a.Email)
	app.WriteJSON(w, http.StatusAccepted, payload)

}
