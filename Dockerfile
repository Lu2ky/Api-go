FROM golang:1.26.0-alpine3.23 AS builder

WORKDIR /app
COPY * .
RUN go mod tidy
RUN go build ./main.go
EXPOSE 3913

FROM alpine:latest
COPY --from=builder /app/main .
RUN chmod u+x main
CMD ["./main"]

