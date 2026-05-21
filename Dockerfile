FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /subscriptions-api .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /subscriptions-api .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["/app/subscriptions-api"]
