# Build stage
FROM golang:1.20-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux go build -ldflags="-s -w"  -o imageUploader ./cmd/

# Final stage
FROM alpine:3.14
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /app/imageUploader .
COPY .env .
EXPOSE 8080
CMD ["./imageUploader"]