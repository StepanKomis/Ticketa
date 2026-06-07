package main

import (
	"log"
	"os"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/cmd/server/startup"
	"github.com/StepanKomis/Ticketa/src/config"
)

func main() {
	cfg, err := config.Load("/config/ticketa.yaml")
	if err != nil {
		log.Fatalf("[FATAL] %s", err)
		os.Exit(1)
	}

	store := config.NewStore(cfg, "/config/ticketa.yaml")

	l, err := logs.NewLogger("server", cfg)
	if err != nil {
		log.Fatalf("[FATAL] %s", err)
		os.Exit(1)
	}

	if err := startup.InitializeServer(l, store); err != nil {
		l.Fatalf("Failed to start server: %s", err)
	}
}
