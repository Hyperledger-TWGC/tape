FROM golang:1.12

WORKDIR /root

COPY . .

RUN go build
