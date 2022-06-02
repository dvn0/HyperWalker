# HyperWalker

> Hypertext grabber...

Snapshot and sanitize webpages just-as-they-are in a headless Firefox.

Using the [Freeze-dry](https://github.com/WebMemex/freeze-dry) javascript library for self-contained HTML file snapshots.

## Roadmap

* Nix Flake (Package Definition)
* Debug mode
* HTTP API
* Make use of freeze-dry's customisation abilities

## Requirements

* Firefox (in your path)
* Go (to compile)

* Earthly (https://earthly.dev - for building)

## Build & Run

```shell
    $ earthly config global.disable_analytics true
    $ earthly config global.disable_log_sharing true
    $ earthly +build
    $ ./build/hyperwalker -url https://en.wikipedia.org/wiki/Special:Random
```

## Test

Warning: This runs Firefox in a docker container and takes several minutes the first time around.

```shell
    $ earthly +firefox-image
    $ earthly +application-test
```
