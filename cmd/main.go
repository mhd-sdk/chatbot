package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mhd-sdk/chatbot/internal/env"
	"github.com/mhd-sdk/chatbot/internal/server"
)

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	slog.Info("Starting Delta service")

	err := env.LoadEnv()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info("Environment variables loaded")

	server := server.New()

	server.ServeAPI()
}
