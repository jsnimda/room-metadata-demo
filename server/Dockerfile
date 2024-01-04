# Use an official Go runtime as a parent image
FROM golang:1.21.3-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files first to leverage Docker cache
COPY go.* ./

# Download all the dependencies
RUN go mod download

# Copy the main application
COPY main.go .

# Expose port 8101 to the outside world
EXPOSE 8101

# Command to run the application
CMD ["go", "run", "main.go"]
