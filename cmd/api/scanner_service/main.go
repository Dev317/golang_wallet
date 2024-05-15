package main

import (
	"log/slog"
	"os"

	cf "github.com/Dev317/golang_wallet/config/scanner"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	config, err := cf.LoadServerConfig(".")
	if err != nil {
		logger.Error("Failed to load config",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	queueConfig, err := cf.LoadQueueConfig(".")
	if err != nil {
		logger.Error("Failed to load queue config",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	server := NewServer(config, queueConfig)
	server.Start()
}