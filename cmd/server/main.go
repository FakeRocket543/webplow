package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"webplow/internal/auth"
	"webplow/internal/config"
	"webplow/internal/handler"
)

func main() {
	cfg := config.Load()

	store, err := auth.NewStore(cfg.TokenFile)
	if err != nil {
		log.Fatalf("load tokens: %v", err)
	}

	os.MkdirAll(cfg.TempDir, 0755)

	h := handler.New(cfg, store)

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Convert)
	mux.HandleFunc("/health", h.Health)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		fmt.Printf("webplow listening on %s\n", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	sighup := make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)
	go func() {
		for range sighup {
			if err := store.Reload(); err != nil {
				log.Printf("token reload failed: %v", err)
			} else {
				log.Println("tokens reloaded")
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
