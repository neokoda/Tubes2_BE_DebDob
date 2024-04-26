# Use the official Golang image for building the app
FROM golang:1.22.2 AS builder

# Set the Current Working Directory inside the container
WORKDIR /usr/app

# Copy the go.mod and go.sum files into the container
COPY go.mod go.sum ./

# Download all Go dependencies (caching these improves build times)
RUN go mod download && go mod verify

# Copy the rest of your Go source code
COPY . ./

RUN go build 

# Expose the port where your app will run (adjust as needed)
EXPOSE 8000

# Command to run the Go app
CMD ["./debdob"]
