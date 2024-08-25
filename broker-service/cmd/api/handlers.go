package main

import (
	"broker/event"
	"broker/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
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
	case "log":
		app.logItemViaRPC(w, requestPayload.Log)
	case "mail":
		app.SendEmail(w, requestPayload.Mail)
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

func (app *App) Log(w http.ResponseWriter, l LogPayload) {
	// create some json we'll send to the log microservice
	jsonData, err := json.MarshalIndent(l, "", "\t")
	if err != nil {
		app.ErrorJSON(w, errors.New("could not marshal json"), http.StatusInternalServerError)
		return
	}

	//call the log service
	url := "http://logger-service:8080/log"

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
	if response.StatusCode != http.StatusAccepted {
		app.ErrorJSON(w, fmt.Errorf("unexpected status code: %d", response.StatusCode), http.StatusInternalServerError)
		return
	}

	var payload responsePayload
	payload.Error = false
	payload.Message = fmt.Sprintf("Logged data: %s", l.Data)
	app.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *App) SendEmail(w http.ResponseWriter, msg MailPayload) {
	jsonData, _ := json.MarshalIndent(msg, "", "\t")

	// call the mail service
	mailServiceURL := "http://mailer-service:8080/send"

	// post to mail service
	request, err := http.NewRequest("POST", mailServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the right status code
	if response.StatusCode != http.StatusAccepted {
		app.ErrorJSON(w, errors.New("error calling mail service"))
		return
	}

	// send back json
	var payload responsePayload
	payload.Error = false
	payload.Message = "Message sent to " + msg.To

	app.WriteJSON(w, http.StatusAccepted, payload)

}

// logEventViaRabbit logs an event using the logger-service. It makes the call by pushing the data to RabbitMQ.
func (app *App) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	var payload responsePayload
	payload.Error = false
	payload.Message = "logged via RabbitMQ"

	app.WriteJSON(w, http.StatusAccepted, payload)
}

// pushToQueue pushes a message into RabbitMQ
func (app *App) pushToQueue(name, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *App) logItemViaRPC(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:5001")
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	rpcPayload := RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}

	var result string
	err = client.Call("RPCServer.LogInfo", rpcPayload, &result)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	payload := responsePayload{
		Error:   false,
		Message: result,
	}

	app.WriteJSON(w, http.StatusAccepted, payload)
}

func (app *App) LogViaGRPC(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.ReadJSON(w, r, &requestPayload)
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}
	defer conn.Close()

	c := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Log.Name,
			Data: requestPayload.Log.Data,
		},
	})
	if err != nil {
		app.ErrorJSON(w, err)
		return
	}

	var payload responsePayload
	payload.Error = false
	payload.Message = "logged"

	app.WriteJSON(w, http.StatusAccepted, payload)
}
