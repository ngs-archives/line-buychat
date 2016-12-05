FROM golang:1.7

MAINTAINER Atsushi Nagase<a@ngs.io>
RUN apt-get update && apt-get -y install libzbar-dev && apt-get clean

COPY vendor /go/src
RUN mkdir -p /go/src/github.com/ngs/line-buychat
WORKDIR /go/src/github.com/ngs/line-buychat
COPY main.go .
COPY app ./app
RUN go build -o /usr/bin/server main.go

CMD /usr/bin/server
