# --- Build Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Use direct as fallback in case proxy.golang.org is flaky
ENV GOPROXY=https://proxy.golang.org,direct
ENV GONOSUMCHECK=*

COPY go.mod go.sum ./
RUN go mod download -x

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /oauth2-provider ./cmd/server

# --- Final Stage ---
FROM alpine:3.19

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /home/appuser

COPY --from=builder /oauth2-provider .

EXPOSE 8080

ENTRYPOINT ["./oauth2-provider"]
