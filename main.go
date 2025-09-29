package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	// This is a wrapper that runs the actual API server
	fmt.Println("Starting OpenTelemetry Example API...")

	// Run the API server
	cmd := exec.Command("go", "run", "./cmd/api/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}
