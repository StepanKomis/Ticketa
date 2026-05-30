package main

import (
	"log"
	"os"

	"github.com/StepanKomis/Ticketa/src/cmd/server/startup"
)

func main() {
	if err := startup.InitializeServer(); err != nil {
		log.Fatalf("Failed to start server: %s", err)
		os.Exit(1)
	}
}
