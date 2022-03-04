# HyperWalker

> Hypertext organizer...

Snapshot and sanitize webpages just-as-they-are using headless Firefox. 

## Roadmap

* Allow passing URI as argument on command line
* Persist the Firefox instance and send it subsequent commands
* HTTP API

## Requirements


* Firefox
* Go

## Install

```shell

    $ go get -v -d git.callpipe.com/dvn/hyperwalker
    $ go get -v github.com/rakyll/statik
    $ cd $GOPATH/src/git.callpipe.com/dvn/hyperwalker
    $ go generate
    $ go build -o hyperwalker main.go
```
