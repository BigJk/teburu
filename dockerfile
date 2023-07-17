# Build stage
FROM golang:1.20-alpine AS build

# Set the working directory to the project root
WORKDIR /go/src/github.com/BigJk/teburu

# Copy the source code to the working directory
COPY . .

# Build the binary
RUN go build -o /go/bin/teburu ./cmd/teburu

# Run stage
FROM alpine:latest

# Copy the binary from the build stage
COPY --from=build /go/bin/teburu /usr/bin/teburu

# Set the entrypoint
ENTRYPOINT ["/usr/bin/teburu"]
