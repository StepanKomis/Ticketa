// Package main je vstupní bod serveru Ticketa.
//
// @title          Ticketa API
// @version        0.1.0
// @description    REST API ticketovacího systému pro školy.
// @description    Studenti a zaměstnanci spravují vlastní tikety, udržovatelé (maintainer) spravují celý systém — uživatele, stavy tiketů a runtime konfiguraci.
// @description
// @description    **Autentizace:** cookie-based. Po přihlášení přes POST /api/login je nastaven HTTP-only cookie `session_token` platný 7 dní.
// @description    Chráněné endpointy vrátí 401 bez platného cookie a 403 při nedostatečné roli.
//
// @contact.name   StepanKomis
// @contact.url    https://github.com/StepanKomis/Ticketa
//
// @host           localhost:8080
// @BasePath       /
//
// @securityDefinitions.apikey  cookieAuth
// @in                          cookie
// @name                        session_token
// @description                 HTTP-only session cookie platná 7 dní. Nastavena automaticky po přihlášení.
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
