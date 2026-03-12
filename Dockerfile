# Stage 1: Build
FROM golang:1.24-alpine AS builder
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001
COPY . .
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# Stage 2: Dev (with Air hot reload)
FROM golang:1.24-alpine AS dev
RUN go install github.com/air-verse/air@latest
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
CMD ["air"]

# Stage 3: Production (~15MB final image)
FROM alpine:3.20 AS production
RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /server /server
COPY --from=builder /app/static /static
COPY --from=builder /app/migrations /migrations
USER app
EXPOSE 8080
CMD ["/server"]
