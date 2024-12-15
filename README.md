# **Go Test Task**

## **Design Decisions**

1. **Simplicity First**:
   - External packages were avoided for the HTTP server (e.g., no `gin`) and database interaction (e.g., no `gorm`) to reduce unnecessary dependencies and complexity.
   - The goal is to keep the build and deployment process straightforward.
   - The REST architecture was chosen because it satisfies the requirements of this task (GET and POST requests, small data payload).

2. **Redis for Idempotency**:
   - Redis is used to check if a `transactionId` has already been processed.
   - This prevents duplicate database updates without storing every transaction ID permanently, as the goal is simply to ensure that each transaction executes only once.

3. **Project Structure**:
   - The project follows a clean structure for scalability and maintainability:
     ```
     testTaskEnt/
     ├── db/                     # Database initialization and queries
     │   ├── db.go
     │   ├── db_test.go
     ├── handlers/               # API request handlers
     │   ├── handlers.go
     │   ├── handlers_test.go
     ├── models/                 # Data models and constants
     │   ├── models.go
     ├── postman/                # Postman collection for API testing
     │   ├── GoTestTask.postman_collection.json
     ├── docker-compose.yml      # Docker Compose configuration
     ├── Dockerfile              # Dockerfile for building the Go service
     ├── go.mod                  # Go module file
     ├── main.go                 # Application entry point
     └── README.md               # Project documentation
     ```

4. **Automated Tests**:
   - Automated 'tests' are included for the database and handlers, but currently they are only fake tests the due to time constraints.

5. **Postman Collection**:
   - A Postman collection is provided in `postman/GoTestTask.postman_collection.json` to test the API.
   - **Workflow**:
     1. **First Test**: Generates a random `transactionId`, processes a transaction, and succeeds.
     2. **Second Test**: Reuses the same `transactionId`, and the request fails due to idempotency.
     3. **Third & Fourth Tests**: Fetch the balance of predefined users.


## **How to Run**

1. Clone the repository:
   ```bash
   git clone https://github.com/mligema/TestTaskEnt.git
   cd GoTestTask
   ```

2. Start the application:
   ```bash
   docker-compose up
   ```

3. Access the API:
   - Balance URL: `http://localhost:8080/user/1/balance`
   - Balance URL: `http://localhost:8080/user/1/transaction` + (POST request needs a body payload too - see postman collection)


## **Predefined Users**

The database includes the following users for testing:

| `userId` | `balance` |
|----------|-----------|
| 1        | 10.15     |
| 2        | 1.15      |
| 3        | 101.15    |


## **Postman Collection**

1. **Import the Collection**:
   - Open Postman.
   - Click **Import** and select `postman/GoTestTask.postman_collection.json`.

2. **Run the Requests**:
   - Click on the GoTestTask collection.
   - Click on three dots and select **Run Collection**, or just click Run and 'Run GoTestTask'.
