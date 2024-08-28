# Use an official Golang image as the base
FROM golang:alpine

# Set the working directory to /app
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Copy the application code
COPY . .

# Run the command to start the application
CMD ["pm2 start 'go run .' --watch '.go' --watch '/.go' --watch '//.go' && pm2 log"]
