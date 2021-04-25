FROM golang:alpine as builder

COPY . /shitpool-core

RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.io,direct

WORKDIR /shitpool-core

RUN go mod tidy
RUN mkdir -p /shitpool
RUN CGO_ENABLED=0 go build -o /shitpool/shitpool /shitpool-core/main.go

FROM alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk update
RUN apk upgrade
RUN apk add --update tzdata
ENV TZ=Asia/Shanghai
RUN rm -rf /var/cache/apk/*

RUN mkdir /shitpool

WORKDIR /shitpool

COPY --from=builder /shitpool/shitpool /shitpool/shitpool

ENV PATH /shitpool/:$PATH

COPY run.sh /shitpool/run.sh

CMD sh run.sh
