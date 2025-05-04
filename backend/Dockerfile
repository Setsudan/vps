FROM golang:1.23.8-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -tags=prod -o main .

FROM alpine:3.21.3

# Create a non‑root user
RUN adduser -D appuser

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=build /app/main ./

# Switch to the non‑root user
USER appuser

EXPOSE 8080
CMD ["./main"]