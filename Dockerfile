# Dockerfile

# Use the official Golang image
FROM golang:1.22.5

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Build all services (you may adjust this depending on your project structure)
RUN go build -o email-service cmd/email/main.go
RUN go build -o recipient-service cmd/recipient/main.go
RUN go build -o template-service cmd/template/main.go
RUN go build -o stats-service cmd/stats/main.go

# Command to run the entire application using supervisor (optional)
CMD ["supervisord", "-c", "/etc/supervisord.conf"]
