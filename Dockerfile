FROM golang:1.19

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o tbd-bot .
EXPOSE 8080
CMD ["./tbd-bot"]
