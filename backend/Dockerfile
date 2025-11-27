# Build stage
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o backend main.go


# Run stage
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/backend .

EXPOSE 8080

CMD ["./backend"]
