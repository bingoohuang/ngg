FROM golang:1.8

ADD . /go/src/github.com/bingoohuang/ngg/gurl

RUN go install github.com/bingoohuang/ngg/gurl

ENTRYPOINT ["/go/bin/bat"]