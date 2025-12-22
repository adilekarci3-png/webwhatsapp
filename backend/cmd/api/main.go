package main

import (
	"log"
	"net/http"
	"time"

	"example.com/webwhatsapp/backend/internal/bootstrap"
)

func main() {
	app, err := bootstrap.Build()
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:              ":" + app.Port,
		Handler:           app.Router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("backend listening on :%s", app.Port)
	log.Fatal(srv.ListenAndServe())
}
