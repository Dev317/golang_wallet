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
)

type Server struct {
	config cf.Config
	q      *db.Queries
	s      *http.Server
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

func NewServer(config cf.Config) *Server {

	q := makeQuery(config)
	s := makeHTTPServer(config)

	server := &Server{
		config: config,
		q:      q,
		s:      s,
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
	}

	logger.Warn("Server shutdown successfully")
}
