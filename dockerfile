FROM golang:1.22-alpine

# Set environment variables
ENV GO111MODULE=on

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application source code into the container
COPY . .

# Install git (required for fetching dependencies)
RUN apk update && \
    apk add --no-cache git && \
    go build -o backup

# Expose the port the application runs on
EXPOSE 8080

# Command to run the application
CMD ["./backup"]
