VERSION 0.5
FROM golang:1.18-alpine3.14
WORKDIR /hyperwalker

deps:
    COPY go.mod go.sum ./
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

freeze-dry:
    FROM node
    #COPY freeze-dry /freeze-dry
    RUN git clone https://github.com/WebMemex/freeze-dry -b customisation /freeze-dry
    WORKDIR /freeze-dry
    RUN ls -alh
    RUN pwd
    RUN npm install && \
        npm run bundle
    SAVE ARTIFACT dist AS LOCAL build/freeze-dry

build:
    FROM +deps
    COPY main.go .
    RUN mkdir ./js/
    COPY +freeze-dry/dist ./js/dist
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
    RUN ./hyperwalker -url https://en.wikipedia.org/wiki/Special:Random ; cat $HOME/.hyperwalker/logs/hyperwalker.log
    RUN ls -alh $HOME/.mozilla/firefox | grep "\.hyperwalker"
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
