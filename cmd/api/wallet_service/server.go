package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cf "github.com/Dev317/golang_wallet/config/wallet"
	db "github.com/Dev317/golang_wallet/db/wallet/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Server struct {
	config    cf.Config
	ethConfig cf.EthereumConfig
	queueConfig cf.QueueConifg
	q         *db.Queries
	s         *http.Server
	queueConn *amqp.Connection
}

func makeQuery(config cf.Config) *db.Queries {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()
	d, err := pgxpool.New(ctx, makeConnString(config))
	if err != nil {
		logger.Error("Failed to create connection pool",
			slog.Any("error", err),
		)
	}
	return db.New(d)
}

func makeHTTPServer(config cf.Config) *http.Server {
	return &http.Server{
		Addr:         config.HTTPServerAddress,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func makeQueueConnection(queueConfig cf.QueueConifg) (*amqp.Connection, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	conn, err := amqp.Dial(queueConfig.URI)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ",
			slog.Any("error", err),
		)
		return nil, err
	}
	return conn, nil
}

func NewServer(config cf.Config, ethConfig cf.EthereumConfig, queueConfig cf.QueueConifg) *Server {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	q := makeQuery(config)
	s := makeHTTPServer(config)

	queueConn, err := makeQueueConnection(queueConfig)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
	logger.Info("Connected to RabbitMQ")

	server := &Server{
		config:    config,
		ethConfig: ethConfig,
		q:         q,
		s:         s,
		queueConn: queueConn,
	}

	mux := http.NewServeMux()
	server.SetupRoutes(mux)
	return server
}

func makeConnString(config cf.Config) string {
	return "user=" + config.DBUser + " password=" + config.DBPassword + " dbname=" + config.DBName + " sslmode=" + config.DBSSLMode + " host=" + config.DBHost + " port=" + config.DBPort
}

func (server *Server) SetupRoutes(mux *http.ServeMux) {
	user := http.NewServeMux()
	user.HandleFunc("/create", server.CreateUser)

	account := http.NewServeMux()
	account.HandleFunc("/create", server.CreateAccount)
	account.HandleFunc("/create_transaction", server.CreateTransaction)

	mux.Handle("/api/v1/user/", http.StripPrefix("/api/v1/user", user))
	mux.Handle("/api/v1/account/", http.StripPrefix("/api/v1/account", account))

	mux.HandleFunc("/api/v1/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("message: Service is healthy!"))
	})

	server.s.Handler = mux
}

func (server *Server) Start() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.s.ListenAndServe(); err != nil {
			logger.Error("Server error",
				slog.Any("error", err),
			)
		}
	}()
	logger.Info("Server started successfully")

	<-done
	logger.Warn("Server stopped!")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.s.Shutdown(ctx); err != nil {
		logger.Error("Failed to shutdown server",
			slog.Any("error", err),
		)
		server.queueConn.Close()
	}

	logger.Warn("Server shutdown successfully")
}
