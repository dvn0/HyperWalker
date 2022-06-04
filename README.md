# HyperWalker

> Hypertext grabber...

Snapshot and sanitize webpages just-as-they-are in a headless Firefox.

Using the [Freeze-dry](https://github.com/WebMemex/freeze-dry) javascript library for self-contained HTML file snapshots.

## Roadmap

* Debug mode
* HTTP API
* Make use of freeze-dry's customisation abilities

## Requirements

### Runtime
* Firefox (in your path)

### Compile
* Go (to compile)
* Earthly *or* Nix (for building)

## Build & Run

***Current Major Bug: HyperWalker will typically segfault the first time it's run. Try running it again, if this happens.***

**Build & Run Manually**

```shell
$ mkdir -p js/dist
$ wget -O js/dist/freeze-dry.umd.js https://git.callpipe.com/dvn/hyperwalker/-/jobs/16551/artifacts/raw/build/freeze-dry/freeze-dry.umd.js
$ go build
$ ./HyperWalker -url https://en.wikipedia.org/wiki/Special:Random
```

**Using Earthly**

```shell
$ earthly config global.disable_analytics true
$ earthly config global.disable_log_sharing true
$ earthly +build
$ ./build/hyperwalker -url https://en.wikipedia.org/wiki/Special:Random
```

**Using Nix Flake**

Nix super-quickstart:

```shell
$ nix run sourcehut:~dvn/HyperWalker -- -url https://en.wikipedia.org/wiki/Special:Random
```

More involved:

```shell
$ nix build
$ mkdir -p $HOME/.hyperwalker/logs
$ ./result/HyperWalker -url https://en.wikipedia.org/wiki/Special:Random

$ # or to build and run directly:
$ mkdir -p $HOME/.hyperwalker/logs
$ nix run . -- -url https://en.wikipedia.org/wiki/Special:Random
```

**Run Using Docker**

```shell
$ docker run --rm registry.git.callpipe.com/dvn/hyperwalker:latest /hyperwalker/hyperwalker -url https://en.wikipedia.org/wiki/Special:Random
```

## Test

Warning: This runs Firefox in a docker container and takes several minutes the first time around.

```shell
$ earthly +firefox-image
$ earthly +application-test
```
