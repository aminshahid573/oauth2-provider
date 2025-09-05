# --- Build Stage ---
# Use the official Go image as a builder.
# Pinning to a specific version ensures reproducible builds.
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application.
# -ldflags="-w -s" strips debug information, reducing the binary size.
# CGO_ENABLED=0 creates a static binary, which is ideal for Alpine images.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /oauth2-provider ./cmd/server

# --- Final Stage ---
# Use a minimal, secure base image. Alpine is a great choice.
FROM alpine:3.19

# It's good practice to run as a non-root user.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Set the working directory
WORKDIR /home/appuser

# Copy the built binary from the builder stage.
COPY --from=builder /oauth2-provider .

# Copy the web directory (templates and static assets) if they are not embedded.
# Since we are using `embed`, this is not strictly necessary but can be useful
# if you ever want to mount external templates. For now, we'll assume embedded.
# COPY web ./web

# Expose the port the application will run on.
EXPOSE 8080

# The command to run when the container starts.
ENTRYPOINT ["./oauth2-provider"]