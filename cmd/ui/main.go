package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/shivamshashank/CloudInferOps/internal/ui"
)

func main() {
	service := ui.NewService()
	handler := ui.NewHandler(service)

	mux := http.NewServeMux()
	mux.Handle("/api/", handler)
	mux.Handle("/", http.FileServer(http.Dir("./web/dist")))

	fmt.Println("CloudInferOps UI server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
