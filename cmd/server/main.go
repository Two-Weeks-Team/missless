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

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/Two-Weeks-Team/missless/internal/auth"
	"github.com/Two-Weeks-Team/missless/internal/config"
	"github.com/Two-Weeks-Team/missless/internal/handler"
	"github.com/Two-Weeks-Team/missless/internal/media"
	"github.com/Two-Weeks-Team/missless/internal/memory"
	"github.com/Two-Weeks-Team/missless/internal/middleware"
	"github.com/Two-Weeks-Team/missless/internal/store"
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

	// Cloud Storage client for uploads (optional — nil if unavailable)
	var storageClient *storage.Client
	if cfg.StorageBucket != "" {
		sc, err := storage.NewClient(ctx)
		if err != nil {
			slog.Warn("storage_client_init_failed", "error", err)
		} else {
			storageClient = sc
			defer sc.Close()
		}
	}
	uploader := media.NewUploader(cfg.StorageBucket, storageClient)

	// Firestore client (optional — nil if unavailable)
	var firestoreClient *firestore.Client
	if cfg.ProjectID != "" {
		fc, err := firestore.NewClientWithDatabase(ctx, cfg.ProjectID, cfg.FirestoreDB)
		if err != nil {
			slog.Warn("firestore_client_init_failed", "error", err)
		} else {
			firestoreClient = fc
			defer fc.Close()
		}
	}

	// maxMemoriesPerPersona is the maximum number of memories stored per persona.
	const maxMemoriesPerPersona = 100

	sessionStore := store.NewFirestoreStore(cfg.ProjectID, firestoreClient)
	memStore := memory.NewStore(maxMemoriesPerPersona, firestoreClient)

	// HTTP mux
	mux := http.NewServeMux()
	sessions := auth.NewSessionStore()
	handler.RegisterHealth(mux)
	handler.RegisterWebSocket(mux, cfg, sessions, sessionStore, memStore)
	handler.RegisterOAuth(mux, cfg, sessions)
	handler.RegisterUpload(mux, uploader)

	// Serve static frontend (Next.js export)
	fs := http.FileServer(http.Dir("web/out"))
	mux.Handle("/", fs)

	// Apply middleware (outermost runs first)
	// Rate limit: 10 req/s per IP, burst 20 — applies to all endpoints
	limiter := middleware.NewIPRateLimiter(10, 20)
	h := middleware.Recovery(middleware.Logging(middleware.RateLimit(limiter)(mux)))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      h,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // WebSocket needs unlimited write
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	util.SafeGo(func() {
		slog.Info("server started", "port", cfg.Port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	})

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
