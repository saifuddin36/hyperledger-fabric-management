# ---- Builder Stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# ---- Final Stage ----
FROM alpine:latest

WORKDIR /app

# Copy only the crypto material needed by main.go
COPY organizations ./organizations

# Copy the compiled binary from the builder stage
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]