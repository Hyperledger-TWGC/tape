FROM alpine as stupid-base

FROM golang:alpine as golang

WORKDIR /root

ENV GOPROXY=https://goproxy.cn,direct
ENV export GOSUMDB=off

COPY . .

RUN go build -v ./cmd/stupid

FROM stupid-base
RUN mkdir -p /config
COPY --from=golang /root/stupid /usr/local/bin

CMD ["stupid"]
