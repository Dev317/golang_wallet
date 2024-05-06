package main

import (
	"os"
	"log/slog"

	cf "github.com/Dev317/golang_wallet/config/wallet"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	config, err := cf.LoadConfig(".")
	if err != nil {
		logger.Error("Failed to load config",
					slog.Any("error", err),
		)
		os.Exit(1)
	}

	server := NewServer(config)
	server.Start()
}
