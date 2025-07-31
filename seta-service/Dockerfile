# build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/server/main.go

# final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .

# The app loads a .env file, so one should be present in the directory
# where the docker-compose command is run, or copied into the image if static.
# COPY .env .

EXPOSE 8080
CMD ["./main"] 