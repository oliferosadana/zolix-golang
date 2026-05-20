package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"zolix/internal/app"
)

func main() {
	app.LoadEnv(".env")

	store, err := buildStore()
	if err != nil {
		log.Fatal(err)
	}
	server := app.NewServer(store, "web/static", "assets")

	addr := ":8080"
	if value := os.Getenv("PORT"); value != "" {
		addr = ":" + value
	}

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           server.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Zolix Shoe Care running at http://localhost%s", addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func buildStore() (app.DataStore, error) {
	if strings.EqualFold(os.Getenv("STORE"), "postgres") {
		return app.NewPostgresStore(app.PostgresConfig{
			DSN: os.Getenv("DATABASE_URL"),
		})
	}
	log.Print("using in-memory store; set STORE=postgres and DATABASE_URL to enable PostgreSQL")
	return app.NewStore(), nil
}
