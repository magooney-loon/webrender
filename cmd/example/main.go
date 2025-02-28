package main

import (
	"context"
	"log"
	"net/http"

	server "github.com/magooney-loon/webserver/pkg/server/impl"
)

func main() {
	// Create a new server
	srv := server.NewServer()

	// Get the router
	r := srv.Router()

	// Register a hello handler
	r.Handle(http.MethodGet, "/hello", helloHandler)

	// Create a v1 group
	v1 := r.Group("/v1")
	// Register a world handler for the group
	v1.Handle(http.MethodGet, "/world", helloHandler)

	// Start the server
	if err := srv.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}
