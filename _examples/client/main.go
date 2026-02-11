// Client example serves the public UI that talks to the JSON:API server.
// Run the server first (e.g. go run ./_examples/server), then run this:
//
//	go run ./_examples/client
//
// Open http://localhost:8081 and use the forms. Validation errors from the
// server are shown next to the relevant fields.
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed public/*
var publicFS embed.FS

func main() {
	root, _ := fs.Sub(publicFS, "public")
	fs := http.FileServer(http.FS(root))
	http.Handle("/", fs)
	log.Println("Client UI at http://localhost:8081 (ensure server is running on :8080)")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
