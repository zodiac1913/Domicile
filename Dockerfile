FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1020

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN templ generate
RUN CGO_ENABLED=0 GOOS=linux go build -o domicile server.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/domicile .
COPY --from=builder /app/static ./static
COPY --from=builder /app/internal/database/schema.sql ./internal/database/schema.sql

EXPOSE 1323
VOLUME /app/data

CMD ["./domicile"]
