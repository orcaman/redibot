FROM golang:latest

ADD . /go/src/github.com/orcaman/redibot

RUN go get github.com/garyburd/redigo/redis
RUN go get github.com/gorilla/websocket
RUN go install github.com/orcaman/redibot

ENTRYPOINT ["/go/bin/redibot"]

EXPOSE 8080