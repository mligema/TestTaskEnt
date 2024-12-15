# Use an official Golang image as the base
FROM golang:1.23

# Set the working directory in the container
WORKDIR /app

# Copy Go modules and dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./main"]