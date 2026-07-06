package main

import (
	"fmt"
	"log"
	"os"

	"github.com/shivamshashank/CloudInferOps/internal/webhook"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Starting Webhook Handler on port %s...\n", port)
	if err := webhook.StartServer(port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
