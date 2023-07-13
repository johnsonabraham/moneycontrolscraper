# syntax=docker/dockerfile:1

FROM golang:1.19 AS base

 WORKDIR /usr/src/app

 COPY go.mod go.sum .env ./

 RUN go mod download && go mod verify

 COPY . .

 FROM base AS moneycontrolscraper
 RUN echo "Running moneycontrolscraper build"
 RUN go build -v -o /usr/local/bin/moneycontrolscraper ./cmd/main.go
 EXPOSE 8080
 CMD [ "moneycontrolscraper" ]