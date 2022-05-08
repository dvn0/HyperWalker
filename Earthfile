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
    FROM registry.git.callpipe.com/dvn/hyperwalker/firefox-image:latest
    WORKDIR /hyperwalker
    RUN groupadd -g 6000 hyperwalker && useradd -r -u 6000 -g hyperwalker -d /home/hyperwalker hyperwalker
    RUN mkdir -p /home/hyperwalker
    RUN chown -R hyperwalker:hyperwalker /hyperwalker /home/hyperwalker
    USER hyperwalker
    COPY +build/hyperwalker .
    RUN mkdir -p $HOME/.hyperwalker/logs
    #ENTRYPOINT ["/hyperwalker/hyperwalker"]
    SAVE IMAGE --push registry.git.callpipe.com/dvn/hyperwalker:latest

application-test:
    FROM +docker
    RUN pwd && ./hyperwalker
    RUN ./hyperwalker -url https://en.wikipedia.org/wiki/Special:Random ; sleep 5 ; ./hyperwalker -url https://en.wikipedia.org/wiki/Special:Random ; ps aux | grep marionette | grep -v grep ; sleep 5 ; ./hyperwalker -url https://en.wikipedia.org/wiki/Special:Random
    RUN cat $HOME/.hyperwalker/logs/hyperwalker.log
    RUN mv /tmp/*hyperwalker*.html test-snapshot.html
    SAVE ARTIFACT test-snapshot.html AS LOCAL test-snapshot.html

firefox-image:
    RUN docker pull registry.git.callpipe.com/dvn/hyperwalker/firefox-image:latest
    FROM DOCKERFILE -f tests/firefox.Dockerfile .
    SAVE IMAGE --push registry.git.callpipe.com/dvn/hyperwalker/firefox-image:latest

all:
  BUILD +build
  BUILD +docker
  BUILD +application-test
