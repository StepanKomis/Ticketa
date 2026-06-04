package main

import (
	"log"
	"os"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/cmd/server/startup"
)

func main() {
	l, err := logs.NewLogger("server")
	if err != nil {
		log.Fatalf("[FATAL] %s", err)
		os.Exit(1)
	}

	if err := startup.InitializeServer(l); err != nil {
		l.Fatalf("Failed to start server: %s", err)
	}
}
