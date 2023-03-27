# build stage
#FROM golang:alpine as build-env
#FROM golang:alpine3.16 as build-env
FROM golang:bullseye as build-env
MAINTAINER mdouchement

RUN apt-get update
RUN apt-get install -y git curl
#RUN apk upgrade
#RUN apk add --update --no-cache git curl

RUN mkdir -p /go/src/github.com/mdouchement/openstackswift
WORKDIR /go/src/github.com/mdouchement/openstackswift

ENV CGO_ENABLED 0
ENV GO111MODULE on
ENV GOPROXY https://proxy.golang.org

COPY . /go/src/github.com/mdouchement/openstackswift
# Dependencies

RUN go version

RUN go mod download

# run tests and save coverage
#RUN CGO_ENABLED=1 go test -cpu 1 -race -coverpkg=./internal/database,./internal/model,./internal/scheduler,./internal/storage,./internal/webserver,./internal/webserver/middleware,./internal/webserver/serializer,./internal/webserver/service,./internal/webserver/weberror,./internal/xpath,./tests -coverprofile=cprof.out ./tests/
RUN go test -cpu 1 -coverpkg=./internal/database,./internal/model,./internal/scheduler,./internal/storage,./internal/webserver,./internal/webserver/middleware,./internal/webserver/serializer,./internal/webserver/service,./internal/webserver/weberror,./internal/xpath,./tests -coverprofile=cprof.out ./tests/
RUN go tool cover -html=cprof.out -o coverage.html

# build
RUN go build -ldflags "-s -w" -o swift ./cmd/swift/main.go

# final stage
FROM debian:bullseye
MAINTAINER mdouchement

ENV DATABASE_PATH /data
ENV STORAGE_PATH /data

RUN mkdir -p ${STORAGE_PATH}

COPY --from=build-env /go/src/github.com/mdouchement/openstackswift/swift /usr/local/bin/
COPY --from=build-env /go/src/github.com/mdouchement/openstackswift/coverage.html /

EXPOSE 5000
CMD ["swift", "server"]
