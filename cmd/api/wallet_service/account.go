package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	db "github.com/Dev317/golang_wallet/db/wallet/sqlc"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type CreateAccountRequest struct {
	UserID  int64 `json:"user_id"`
	ChainID int32 `json:"chain_id"`
}

func createAddress() (string, string, string, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		logger.Error("Failed to generate private key",
			slog.Any("error", err),
		)
		return "", "", "", err
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyStr := hexutil.Encode(privateKeyBytes)[2:]
	logger.Warn("Private key generated", slog.String("private_key", privateKeyStr))

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logger.Error("Failed to assert type: publicKey is not of type *ecdsa.PublicKey")
		return "", "", "", err
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	pubKeyStr := hexutil.Encode(publicKeyBytes)[4:]
	logger.Warn("Public key generated", slog.String("public_key", pubKeyStr))

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	logger.Warn("Address generated", slog.String("address", address))

	return address, pubKeyStr, privateKeyStr, nil
}

func (server *Server) CreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	newAccount := &CreateAccountRequest{}
	err := json.NewDecoder(r.Body).Decode(newAccount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	address, pubKey, privateKey, err := createAddress()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}


	_, err = server.q.CreateAccount(r.Context(), db.CreateAccountParams{
		UserID:  newAccount.UserID,
		ChainID: newAccount.ChainID,
		Address: address,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Account created successfully!\nAddress: " + address + "\nPublic Key: " + pubKey + "\nPrivate Key: " + privateKey))
}
