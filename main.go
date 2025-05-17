package main

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/christiansoetanto/tbd-bot/config"
	"github.com/christiansoetanto/tbd-bot/database"
	"github.com/christiansoetanto/tbd-bot/dbot"
	"github.com/christiansoetanto/tbd-bot/dbot/handler"
	"github.com/christiansoetanto/tbd-bot/logv2"
	"github.com/christiansoetanto/tbd-bot/provider"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

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

	defer func() {
		e := dbotObject.CloseDiscordgoConn()
		if e != nil {
			logv2.Debug(ctx, logv2.Warning, "discordgo connection closed with error: "+e.Error())
		} else {
			logv2.Debug(ctx, logv2.Info, "discordgo connection closed")
		}
		database.Close(ctx)
	}()

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	//prov.HelloWorld(ctx)
	// Wait here until CTRL-C or other term signal is received.
	logv2.Debug(ctx, logv2.Info, "Session is now running.  Press CTRL-C to exit.")
	// Start HTTP server in a goroutine
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for Azure App Service
	}

	go func() {
		logv2.Debug(ctx, logv2.Info, fmt.Sprintf("Starting HTTP server on port %s", port))
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			logv2.Error(ctx, err, "HTTP server failed")
		}
	}()
	sc := make(chan os.Signal, 1)
	//syscall.SIGTERM,
	signal.Notify(sc, syscall.SIGINT)
	<-sc

	logv2.Debug(ctx, logv2.Info, "Gracefully shutting down.")
}
