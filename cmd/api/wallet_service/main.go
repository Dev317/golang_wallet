package main

import (
	"context"
	"log"
	"net/http"

	db "github.com/Dev317/golang_wallet/db/wallet/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	d, err := pgxpool.New(ctx, "user=postgres password=postgres dbname=wallet sslmode=disable host=localhost port=5432")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	q := db.New(d)
	mux := http.NewServeMux()
	server := NewServer(q, mux)
	err = server.Start()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	log.Println("Server started successfully")

}
