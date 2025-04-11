FROM golang:1.23 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o go-deps-mcp

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/go-deps-mcp .

# Run in HTTP mode by default
ENV PORT=8080
EXPOSE ${PORT}

CMD ["/app/go-deps-mcp", "-http", "-port", "${PORT}"]
