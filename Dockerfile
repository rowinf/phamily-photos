# Start with a base Go image
FROM golang:1.23.3

# Set the working directory
WORKDIR /phamily-photos

# Copy Go modules and dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o phamily-photos .

# Expose the port the app runs on
EXPOSE 8080

CMD ["./phamily-photos"]
