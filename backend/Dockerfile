# Build stage
# Pull golang from mirror
FROM registry.hub.docker.com/library/golang:1.22 AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o backend

# Run stage
FROM registry.hub.docker.com/library/alpine:3.20
WORKDIR /app
COPY --from=builder /app/backend .
EXPOSE 8080
CMD ["./backend"]

