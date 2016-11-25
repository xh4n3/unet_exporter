FROM golang:1.7.3

RUN ["go", "get", "github.com/xh4n3/unet_exporter"]

WORKDIR /go/src/github.com/xh4n3/unet_exporter

ENTRYPOINT ["go", "run", "main.go"]