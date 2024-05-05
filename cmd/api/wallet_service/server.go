package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	db "github.com/Dev317/golang_wallet/db/wallet/sqlc"
)

type Server struct {
	q *db.Queries
	s *http.Server
}

func NewServer(q *db.Queries, mux *http.ServeMux) *Server {
	server := &Server{
		q: q,
		s: &http.Server{
			Addr:         ":8080",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 90 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
	server.SetupRoutes(mux)
	return server
}

func (server *Server) SetupRoutes(mux *http.ServeMux) {
	user := http.NewServeMux()
	user.HandleFunc("/create", server.CreateUser)

	// account := http.NewServeMux()

	mux.Handle("/user/", http.StripPrefix("/user", user))
	// mux.Handle("api/v1/account/", http.StripPrefix("api/v1/account", account))

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello!\n"))
	})

	server.s.Handler = mux
}

func (server *Server) Start() error {
	err := server.s.ListenAndServe()
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
	return nil
}
