# TinyGo WasmServe

An HTTP server for TinyGo Wasm testing like `gopherjs serve`

Fork of [hajimehoshi/wasmserve][https://github.com/hajimehoshi/wasmserve]

## Installation

```sh
go get -u github.com/dImrich/tinygo-wasmserve
```

## Usage

```
Usage of tinygo-wasmserve
  -allow-origin string
        Allow specified origin (or * for all origins) to make requests to this server
  -http string
        HTTP bind address to serve (default ":8080")
  -tags string
        Build tags
  -no-debug bool
        Disable outputting debug symbols. Avoiding debug symbols can have a big impact on generated binary size, reducing them by more than half.
```

## Example

Running a remote package

```sh
# Be careful that `-tags=example` is required to run the below example application.
tinygo-wasmserve -tags=example github.com/dImrich/tinygo-wasmserve/example
```

And open `http://localhost:8080/` on your browser.

## Example 2

Running a local package

```sh
git clone https://github.com/hajimehoshi/ebiten # This might take several minutes.
cd ebiten
tinygo-wasmserve -tags=example ./examples/sprites
```

And open `http://localhost:8080/` on your browser.

## Known issue with Windows Subsystem for Linux (WSL)

This application sometimes does not work under WSL, due to bugs in WSL, see https://github.com/hajimehoshi/wasmserve/issues/5 for details.

[https://github.com/hajimehoshi/wasmserve]: https://github.com/hajimehoshi/wasmserve