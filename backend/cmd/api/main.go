// Command api is the entry point of the marketplace REST API.
//
// main.go stays deliberately thin: it only *wires things together*
// (config -> database -> routes -> HTTP server). All real logic lives in
// packages under internal/, which keeps it testable without starting a server.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/kthesword/marketplace/backend/internal/config"
	"github.com/kthesword/marketplace/backend/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("%v", err) // prints the message and exits with status 1
	}

	if cfg.MigrateOnStart {
		if err := db.Migrate(cfg.DatabaseURL); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		log.Println("migrations applied")
	}

	// ctx is cancelled when the process receives Ctrl+C (SIGINT) or SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           newRouter(cfg, pool),
		ReadHeaderTimeout: 5 * time.Second, // timeouts defend against slow/malicious clients
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Run the server in a goroutine so main can wait for the shutdown signal.
	serverErr := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", srv.Addr)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr: // server failed on its own (e.g. port already in use)
		log.Fatalf("server: %v", err)
	case <-ctx.Done(): // shutdown requested
		log.Println("shutting down...")
	}

	// Give in-flight requests up to 10s to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("shutdown: %v", err)
	}
	log.Println("bye")
}
