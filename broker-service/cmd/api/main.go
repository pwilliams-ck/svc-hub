package main

import (
	"fmt"
	"log"
	"net/http"
)

const port = "80"

type application struct{}

func main() {
	app := application{}

	log.Printf("Server starting on port %s", port)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}
