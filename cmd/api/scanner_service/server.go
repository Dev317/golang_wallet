package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cf "github.com/Dev317/golang_wallet/config/scanner"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Server struct {
	config    cf.Config
	queueConfig cf.QueueConfig
	s         *http.Server
	queueConn *amqp.Connection
}

func makeHTTPServer(config cf.Config) *http.Server {
	return &http.Server{
		Addr:         config.HTTPServerAddress,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func makeQueueConnection(queueConfig cf.QueueConfig) (*amqp.Connection, error) {
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

func NewServer(config cf.Config, queueConfig cf.QueueConfig) *Server {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
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
		s:         s,
		queueConn: queueConn,
	}

	mux := http.NewServeMux()
	server.SetupRoutes(mux)
	return server
}

func (server *Server) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("message: Service is healthy!"))
	})

	server.s.Handler = mux
}

func (server *Server) Consume(queueName string) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ch, err := server.queueConn.Channel()
	if err != nil {
		logger.Error("Failed to open a channel",
			slog.Any("error", err),
		)
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	if err != nil {
		logger.Error("Failed to declare a queue",
			slog.Any("error", err),
		)
		return
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		logger.Error("Failed to consume messages",
			slog.Any("error", err),
		)
		return
	}

	var forever chan struct{}

	go func() {
		for d := range msgs {
			logger.Info("Received a message",
				slog.Any("message", string(d.Body)),
			)
		}
	}()

	logger.Info(" [*] Waiting for messages")
	<-forever
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

	go server.Consume("scan_queue")

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
