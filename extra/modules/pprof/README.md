## Middleware information

This folder contains a middleware which enables net/http/pprof.


## Install

```sh
$ go get -u github.com/kenorld/egret-contrib/middleware/pprof
```

## Usage

```go
package main

import (
	"github.com/kenorld/egret-contrib/middleware/pprof"
	"github.com/kenorld/egret"
)

func main() {
  
	egret.Get("/", func(ctx *egret.Context) {
		ctx.HTML(egret.StatusOK, "<h1> Please click <a href='/debug/pprof'>here</a>")
	})

	egret.Get("/debug/pprof/*action", pprof.New())

	egret.Listen(":8080")
}

```
