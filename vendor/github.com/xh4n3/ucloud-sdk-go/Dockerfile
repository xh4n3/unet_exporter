FROM golang:1.5

# TODO: Vendor these `go get` commands using Godep.
RUN go get github.com/tools/godep

ENV GOPATH /go/src/github.com/xh4n3/ucloud-sdk-go/Godeps/_workspace:/go
ENV USER root

WORKDIR /go/src/github.com/xh4n3/ucloud-sdk-go

ADD . /go/src/github.com/xh4n3/ucloud-sdk-go
