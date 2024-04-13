# First stage: build stage
FROM golang:1.22-alpine as build

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

# Second stage: final stage
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the build stage to the final stage
COPY --from=build /app/backup .

# Expose the port the application runs on
EXPOSE 8080

# Command to run the application
CMD ["./backup"]
