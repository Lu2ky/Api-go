FROM golang:1.26.2-alpine3.23 AS builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .


FROM alpine:latest
COPY --from=builder /app/main .
RUN chmod u+x main
RUN touch .env
EXPOSE 3913
CMD ["./main"]

