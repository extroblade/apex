package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"apex/internal/config"
	"apex/internal/db"
	"apex/internal/migrate"
	"apex/internal/racing"
	"apex/internal/server"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer database.Close()

	if err := migrate.Run(database); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// Seed the real catalog + season schedule (bundled from the official data)
	// so the planner works without any external source.
	if err := racing.SeedCatalog(context.Background(), database); err != nil {
		log.Fatalf("seed catalog: %v", err)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      server.New(cfg, database),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("bye")
}
