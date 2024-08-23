package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "8080"

type App struct{}

func main() {

	app := App{}

	log.Printf("Strating broker service on port %s\n", webPort)

	//setup the server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}
