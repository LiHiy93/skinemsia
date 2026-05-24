package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"skinemsia/internal/api"
	"skinemsia/internal/bot"
	"skinemsia/internal/config"
	"skinemsia/internal/db"
	"skinemsia/internal/store"
)

func main() {
	cfg := config.Load()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(pool); err != nil {
		log.Fatalf("db migrate: %v", err)
	}
	log.Println("Migrations applied")

	st := store.New(pool)
	handler := api.NewServer(cfg, st)

	// start bot in background (optional — bot may be absent in dev)
	if cfg.BotToken != "" {
		tgBot, err := bot.New(cfg.BotToken, st, cfg.WebAppURL)
		if err != nil {
			log.Printf("WARNING: bot init failed: %v", err)
		} else {
			go tgBot.Start()
			defer tgBot.Stop()
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("Server listening on :%s (env=%s)", cfg.Port, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Server stopped")
}
