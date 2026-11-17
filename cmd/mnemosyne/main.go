package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/profmoo/mnemosyne/db"
	"github.com/profmoo/mnemosyne/internal/server"
)

func main() {
	database, err := db.Open(context.Background(), "mnemosyne.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           server.NewHandler(database),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
