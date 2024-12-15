package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testTaskEnt/db"
	"testTaskEnt/handlers"
)

func main() {
	// Load environment variables for database
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Load environment variables for Redis
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	// Construct PostgreSQL connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	// Initialize the database and open a connection
	database, err := db.InitializeDatabase(connStr)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	// Ensure the database connection is closed when the application exits
	defer database.Close()

	// Initialize the Redis client
	redisClient := handlers.NewRedisClient(redisHost, redisPort)

	// Create an instance of the App struct, passing the database connection to pointer to handlers.App
	app := &handlers.App{
		DB:    database,
		Redis: redisClient,
	}

	// Set up the HTTP server to handle routes under "/user/"
	http.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {

		// Trim the "/user/" prefix from the URL path
		// because we only care about the user ID and endpoint
		path := strings.TrimPrefix(r.URL.Path, "/user/")

		// Split the remaining path into two parts: user ID and endpoint
		pathParts := strings.SplitN(path, "/", 2)

		// Validate URL is 2 parts.
		if len(pathParts) != 2 {
			http.Error(w, "Invalid route", http.StatusNotFound)
			return
		}

		// Extract user ID and endpoint from the split parts
		userIdStr, endpoint := pathParts[0], pathParts[1]

		// Convert the user ID string to a uint64
		userId, err := strconv.ParseUint(userIdStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Route the request based on the endpoint and HTTP method
		switch {
		case endpoint == "transaction" && r.Method == http.MethodPost:
			app.HandleTransaction(w, r, userId)
		case endpoint == "balance" && r.Method == http.MethodGet:
			app.HandleBalance(w, r, userId)
		default:
			http.Error(w, "Invalid endpoint or method", http.StatusNotFound)
		}
	})

	log.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
