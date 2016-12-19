FROM golang:1.7.3

RUN echo 'deb http://archive.ubuntu.com/ubuntu/ trusty universe' > /etc/apt/sources.list

RUN apt-key adv --recv-keys --keyserver hkp://keyserver.ubuntu.com:80 3B4FE6ACC0B21F32

RUN apt-key adv --recv-keys --keyserver hkp://keyserver.ubuntu.com:80 40976EAF437D05B5

RUN apt-get update

RUN apt-get install -y jq

RUN ["go", "get", "github.com/xh4n3/unet_exporter"]

WORKDIR /go/src/github.com/xh4n3/unet_exporter

RUN ["go", "build"]

ENTRYPOINT ["unet_exporter"]