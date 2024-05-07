package main

import (
	"log/slog"
	"os"

	cf "github.com/Dev317/golang_wallet/config/wallet"
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

	ethConfig, err := cf.LoadEthereumConfig(".")
	if err != nil {
		logger.Error("Failed to load ethereum config",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	server := NewServer(config, ethConfig)
	server.Start()
}
