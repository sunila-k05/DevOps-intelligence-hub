# Build stage
FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./main.go

# Runtime stage
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=builder /app/server /app/server
ENV PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
