# syntax=docker/dockerfile:1.4

FROM golang:1.20-alpine AS builder

WORKDIR /rislive

ENV CGO_ENABLED 0
ENV GOPATH /go
ENV GOCACHE /go-build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o rislive

EXPOSE 8054

CMD ["/rislive/rislive"]
