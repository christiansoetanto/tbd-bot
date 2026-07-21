package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/database"
	"github.com/christiansoetanto/tbd-bot/dbot"
	"github.com/christiansoetanto/tbd-bot/dbot/handler"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/provider"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type DiscordCloser interface {
	Close() error
}

func main() {
	ctx := context.Background()
	logv2.Debug(ctx, logv2.Info, "Starting tbd-bot...")
	devMode, err := strconv.ParseBool(os.Getenv("DEVMODE"))
	if err != nil {
		log.Fatal("Error parsing DEVMODE environment variable")
		return
	}

	cfg := config.Init(devMode)

	logv2.Init(cfg.AppConfig)
	session, err := discordgo.New(fmt.Sprintf("Bot %s", os.Getenv("BOTTOKEN")))
	if err != nil {
		log.Fatal(err)
	}

	prov := provider.GetProvider(&provider.Resource{
		AppConfig: cfg.AppConfig,
		Database:  database.GetDBObject(ctx, cfg.AppConfig),
	})

	handlerResource := &handler.Resource{
		Config:   cfg,
		Provider: prov,
	}
	handlerObj := handler.GetHandler(handlerResource)

	dbotResource := &dbot.Resource{
		Config:  cfg,
		Session: session,
		Handler: handlerObj,
	}

	dbotObject := dbot.GetUsecaseObject(dbotResource)
	err = dbotObject.Init(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer database.Close(ctx)

	handler := setupRoutes()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	go func() {
		logv2.Debug(ctx, logv2.Info, fmt.Sprintf("Starting HTTP server on port %s", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logv2.Error(ctx, err, "HTTP server failed")
		}
	}()

	logv2.Debug(ctx, logv2.Info, "Session is now running. Press CTRL-C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	if err := gracefulShutdown(ctx, srv, session, sc); err != nil {
		logv2.Error(ctx, err, "Error during graceful shutdown")
	}
}

func gracefulShutdown(ctx context.Context, srv *http.Server, session DiscordCloser, sigChan <-chan os.Signal) error {
	sig := <-sigChan
	logv2.Debug(ctx, logv2.Info, fmt.Sprintf("Received signal %v, shutting down gracefully...", sig))

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var firstErr error

	if srv != nil {
		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logv2.Error(ctx, err, "HTTP server shutdown error")
			firstErr = err
		} else {
			logv2.Debug(ctx, logv2.Info, "HTTP server shut down successfully")
		}
	}

	if session != nil {
		if err := session.Close(); err != nil {
			logv2.Error(ctx, err, "Discord session close error")
			if firstErr == nil {
				firstErr = err
			}
		} else {
			logv2.Debug(ctx, logv2.Info, "Discord session closed successfully")
		}
	}

	return firstErr
}

func setupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK /health"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK /"))
	})
	return mux
}

