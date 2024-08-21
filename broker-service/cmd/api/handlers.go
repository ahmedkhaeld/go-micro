package main

import (
	"encoding/json"
	"net/http"
)

type response struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (app *App) Broker(w http.ResponseWriter, r *http.Request) {
	payload := response{
		Error:   false,
		Message: "Hit the broker",
	}

	resp, _ := json.MarshalIndent(payload, "", "\t")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(resp)
}
