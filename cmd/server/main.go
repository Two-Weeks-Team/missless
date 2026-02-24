package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Two-Weeks-Team/missless/internal/auth"
	"github.com/Two-Weeks-Team/missless/internal/config"
	"github.com/Two-Weeks-Team/missless/internal/handler"
	"github.com/Two-Weeks-Team/missless/internal/middleware"
	"github.com/Two-Weeks-Team/missless/internal/util"
)

func main() {
	// Structured logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	// Signal context
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// HTTP mux
	mux := http.NewServeMux()
	sessions := auth.NewSessionStore()
	handler.RegisterHealth(mux)
	handler.RegisterWebSocket(mux, cfg)
	handler.RegisterOAuth(mux, cfg, sessions)
	handler.RegisterUpload(mux, cfg)

	// Serve static frontend (Next.js export)
	fs := http.FileServer(http.Dir("web/out"))
	mux.Handle("/", fs)

	// Apply middleware
	h := middleware.Recovery(middleware.Logging(mux))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // WebSocket needs unlimited write
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	go func() {
		slog.Info("server started", "port", cfg.Port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	slog.Info("shutdown signal received")

	// Graceful shutdown (Cloud Run: 8s budget)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	util.SafeGo(func() {
		defer wg.Done()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("http shutdown error", "error", err)
		}
	})
	wg.Wait()

	slog.Info("graceful shutdown complete")
}
