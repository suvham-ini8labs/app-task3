# builder
FROM golang:1.26.0 as builder

WORKDIR /app

COPY . .

RUN adduser build-user --disabled-password && \
    chown -R build-user:build-user /app

USER build-user

RUN go mod download && go mod verify

ENV SERVER_PORT=8080

RUN CGO_ENABLED=0 go build -o ./order-service ./cmd/server/main.go

#runner
FROM alpine:3.23 

COPY --from=builder /app/order-service /order-service

RUN adduser run-user --disabled-password && \
    chown run-user:run-user /order-service

USER run-user

EXPOSE 8080

CMD ["/order-service"]

#