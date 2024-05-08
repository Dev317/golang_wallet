package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"time"

	db "github.com/Dev317/golang_wallet/db/wallet/sqlc"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	amqp "github.com/rabbitmq/amqp091-go"
)

type CreateAccountRequest struct {
	UserID  int64 `json:"user_id"`
	ChainID int32 `json:"chain_id"`
}

type CreateAccountResponse struct {
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Messsage   string `json:"message"`
}

type CreateTransactionRequest struct {
	AccountId  int64  `json:"account_id"`
	ToAddress  string `json:"to_address"`
	Amount     int64  `json:"amount"`
	ChainId    string `json:"chain_id"`
	PrivateKey string `json:"private_key"`
}

type CreateTransactionResponse struct {
	Messsage        string `json:"message"`
	TransactionHash string `json:"transaction_hash"`
	ToAddress       string `json:"to_address"`
	Status          string `json:"status"`
}

type TransactionEvent struct {
	TransactionHash string `json:"transaction_hash"`
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          int64  `json:"amount"`
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
	response := &CreateAccountResponse{
		Messsage:   "Account created successfully!",
		Address:    address,
		PublicKey:  pubKey,
		PrivateKey: privateKey,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(*response)
}

func makeTransaction(pk string, fromHexAddress string, toHexAddress string, value *big.Int, client *ethclient.Client) (string, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	transactionHash := ""
	err := error(nil)

	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		logger.Error("Error converting hex to ECDSA", slog.Any("error", err))
		return transactionHash, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logger.Error("Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		return transactionHash, err
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	if fromAddress.Hex() != fromHexAddress {
		logger.Error(
			"Address mismatch",
			slog.String("from_address", fromAddress.Hex()),
			slog.String("from_hex_address", fromHexAddress),
		)
		return transactionHash, err
	}

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		logger.Error("Error in getting nonce", slog.Any("error", err))
		return transactionHash, err
	}

	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		logger.Error("Error in getting gas price", slog.Any("error", err))
		return transactionHash, err
	}

	toAddress := common.HexToAddress(toHexAddress)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		logger.Error("Error in getting network ID", slog.Any("error", err))
		return transactionHash, err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		logger.Error("Error in signing transaction", slog.Any("error", err))
		return transactionHash, err
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		logger.Error("Error in sending transaction", slog.Any("error", err))
		return transactionHash, err
	}

	txHash := signedTx.Hash().Hex()
	logger.Info("Transaction hash", slog.String("tx_hash", txHash))
	return txHash, err
}

func (server *Server) emitTransactionEvent(queueName string, event *TransactionEvent) {
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
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		logger.Error("Failed to declare a queue",
			slog.Any("error", err),
		)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal event",
			slog.Any("error", err),
		)
		return
	}

	err = ch.PublishWithContext(
		ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
            ContentType: "application/json",
			Body:        []byte(body),
		})
	if err != nil {
		logger.Error("Failed to publish a message",
			slog.Any("error", err),
		)
		return
	}
	logger.Info("Event published successfully")
}

func (server *Server) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	newTransaction := &CreateTransactionRequest{}

	err := json.NewDecoder(r.Body).Decode(newTransaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the chain URL
	var chainURL string
	for _, chainItem := range server.ethConfig.ChainItemList {
		if chainItem.ChainID == newTransaction.ChainId {
			chainURL = chainItem.RPCURL
		}
	}

	client, err := ethclient.Dial(chainURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fromHexAddress, err := server.q.GetAccountAddressById(r.Context(), newTransaction.AccountId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	transactionHash, err := makeTransaction(newTransaction.PrivateKey, fromHexAddress, newTransaction.ToAddress, big.NewInt(newTransaction.Amount), client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	response := &CreateTransactionResponse{
		Messsage:        "Transaction created!",
		TransactionHash: transactionHash,
		ToAddress:       newTransaction.ToAddress,
		Status:          "pending_confirmation",
	}

	event := &TransactionEvent{
		TransactionHash: transactionHash,
		FromAddress:     fromHexAddress,
		ToAddress:       newTransaction.ToAddress,
		Amount:          newTransaction.Amount,
	}

	server.emitTransactionEvent("scan_queue", event)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(*response)
}
