# build stage
FROM golang:bullseye as build-env
MAINTAINER mdouchement

RUN apt-get update
RUN apt-get install -y git curl

RUN mkdir -p /go/src/github.com/mdouchement/openstackswift
WORKDIR /go/src/github.com/mdouchement/openstackswift

ENV CGO_ENABLED 0
ENV GO111MODULE on
ENV GOPROXY https://proxy.golang.org

COPY . /go/src/github.com/mdouchement/openstackswift

# Dependencies
RUN go mod download

# build
RUN go build -ldflags "-s -w" -o swift ./cmd/swift/main.go

# final stage
FROM debian:bullseye
MAINTAINER mdouchement

ENV DATABASE_PATH /data
ENV STORAGE_PATH /data

RUN mkdir -p ${STORAGE_PATH}

COPY --from=build-env /go/src/github.com/mdouchement/openstackswift/swift /usr/local/bin/

EXPOSE 5000
CMD ["swift", "server"]
