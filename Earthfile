VERSION 0.5
FROM golang:1.17-alpine3.14
WORKDIR /hyperwalker

deps:
    COPY go.mod go.sum ./
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

build:
    FROM +deps
    COPY main.go .
    COPY js ./js/
    RUN CGO_ENABLED=0 go build -o build/hyperwalker main.go
    SAVE ARTIFACT build/hyperwalker AS LOCAL build/hyperwalker

docker:
    COPY +build/hyperwalker .
    ENTRYPOINT ["/hyperwalker/hyperwalker"]
    SAVE IMAGE hyperwalker:latest

all:
  BUILD +build
#  BUILD +unit-test
#  BUILD +integration-test
  BUILD +docker
