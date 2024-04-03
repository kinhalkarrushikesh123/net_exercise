FROM golang:1.22

# working directory inside the container
WORKDIR /app

# Copy the Go application source code into the container
COPY . .

# Build the Go application
RUN go build -o backup

# Expose the port the application runs on
EXPOSE 8080

# Command to run the application
CMD ["./backup"]