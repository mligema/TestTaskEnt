package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"testTaskEnt/models"
)

// Define the maximum user ID
const MaxUserID = 99999999

// App struct to hold the database connection
type App struct {
	DB    *sql.DB
	Redis *redis.Client
}

var ctx = context.Background()

// Initialize Redis client
func NewRedisClient(host, port string) *redis.Client {
	addr := fmt.Sprintf("%s:%s", host, port)
	return redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0, // Default DB
	})
}

// HandleTransaction processes a transaction for a specific user
func (app *App) HandleTransaction(w http.ResponseWriter, r *http.Request, userId uint64) {
	var tx models.Transaction

	// Parse JSON payload
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if tx.TransactionId == "" || tx.State == "" || tx.Amount == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Validate transaction state
	if !models.ValidStates[tx.State] {
		http.Error(w, "Invalid transaction state", http.StatusBadRequest)
		return
	}

	// Validate source type
	sourceType := r.Header.Get("Source-Type")
	if !models.ValidSourceTypes[sourceType] {
		http.Error(w, "Invalid Source-Type", http.StatusBadRequest)
		return
	}

	// Validate amount
	amount, err := strconv.ParseFloat(tx.Amount, 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	// Validate user_id
	// Check if userId is valid (positive and within range)
	if userId < 0 || userId > MaxUserID {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	// Process the transaction
	err = app.ProcessTransaction(userId, tx, sourceType)
	if err != nil {
		handleTransactionError(w, err)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Transaction processed successfully"))
}

// HandleBalance retrieves the balance for a specific user
func (app *App) HandleBalance(w http.ResponseWriter, r *http.Request, userId uint64) {
	var balance float64

	// Query user balance from the database
	err := app.DB.QueryRow("SELECT balance FROM users WHERE user_id = $1", userId).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch balance", http.StatusInternalServerError)
		}
		return
	}

	// Construct and send the response
	response := map[string]interface{}{
		"userId":  userId,
		"balance": fmt.Sprintf("%.2f", balance),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) ProcessTransaction(userId uint64, tx models.Transaction, sourceType string) error {
	// Check if transaction ID already exists in Redis
	exists, err := app.Redis.Exists(ctx, tx.TransactionId).Result()
	if err != nil {
		return fmt.Errorf("Failed to check transaction ID: %v", err)
	}
	if exists > 0 {
		return fmt.Errorf("Transaction already processed")
	}

	// Begin a new database transaction
	txDb, err := app.DB.Begin()
	if err != nil {
		return fmt.Errorf("Failed to start transaction: %v", err)
	}
	defer func() {
		// Rollback if something goes wrong
		if err != nil {
			txDb.Rollback()
		}
	}()

	// Lock user balance with SELECT ... FOR UPDATE
	var balance float64
	err = txDb.QueryRow("SELECT balance FROM users WHERE user_id = $1 FOR UPDATE", userId).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("User not found")
		}
		return fmt.Errorf("Failed to fetch user balance: %v", err)
	}

	// Calculate new balance
	amount, _ := strconv.ParseFloat(tx.Amount, 64)
	if tx.State == "win" {
		balance += amount
	} else if tx.State == "lose" {
		if balance < amount {
			return fmt.Errorf("Insufficient balance")
		}
		balance -= amount
	} else {
		return fmt.Errorf("Invalid transaction state")
	}

	// Update balance in SQL
	_, err = txDb.Exec("UPDATE users SET balance = $1 WHERE user_id = $2", balance, userId)
	if err != nil {
		return fmt.Errorf("Failed to update balance: %v", err)
	}

	// Save new transaction in Redis
	transactionData, _ := json.Marshal(tx)
	err = app.Redis.Set(ctx, tx.TransactionId, transactionData, 0).Err()
	if err != nil {
		return fmt.Errorf("Failed to save transaction in Redis: %v", err)
	}

	// Commit the transaction
	err = txDb.Commit()
	if err != nil {
		return fmt.Errorf("Failed to commit transaction: %v", err)
	}

	return nil
}

// handleTransactionError handles errors during transaction processing
func handleTransactionError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		// ErrNoRows, User not found
		http.Error(w, "User not found", http.StatusNotFound)
	} else if strings.Contains(err.Error(), "Transaction already processed") {
		// Transaction already has happened
		http.Error(w, "Transaction already processed", http.StatusConflict)
	} else {
		// I dont know, some other error
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
