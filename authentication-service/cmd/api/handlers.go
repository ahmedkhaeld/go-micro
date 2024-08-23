package main

import (
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

	payload := response{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.WriteJSON(w, http.StatusAccepted, payload)

}
