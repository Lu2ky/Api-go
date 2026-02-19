FROM ubuntu:latest

WORKDIR /app
COPY * .
RUN chmod +x main
RUN touch .env
EXPOSE 3913

CMD ["./main"]

