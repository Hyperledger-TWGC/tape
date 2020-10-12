FROM alpine as tape-base

FROM golang:alpine as golang

WORKDIR /root

ENV GOPROXY=https://goproxy.cn,direct
ENV export GOSUMDB=off

COPY . .

RUN go build -v ./cmd/tape

FROM tape-base
RUN mkdir -p /config
COPY --from=golang /root/tape /usr/local/bin

CMD ["tape"]
