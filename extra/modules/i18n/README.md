## Middleware information

This folder contains a middleware for internationalization uses a third-party package named i81n.

More can be found here:
[https://github.com/Unknwon/i18n](https://github.com/Unknwon/i18n)

## Install

```sh
$ go get -u github.com/egret-contrib/middleware/i18n
```

## Description

Package i18n is for app Internationalization and Localization.


## How to use

Create folder named 'locales'
```
///Files:

./locales/locale_en-US.ini
./locales/locale_el-US.ini
```
Contents on locale_en-US:
```
hi = hello, %s
```
Contents on locale_el-GR:
```
hi = ����, %s
```

```go

package main

import (
	"github.com/kenorld/egret"
	"github.com/egret-contrib/middleware/i18n"
)

func main() {

	egret.UseFunc(i18n.New(i18n.Config{Default: "en-US",
		Languages: map[string]string{
			"en-US": "./locales/locale_en-US.ini",
			"el-GR": "./locales/locale_el-GR.ini",
			"zh-CN": "./locales/locale_zh-CN.ini"}}))
	// or egret.Use(i18n.I18nHandler(....))
	// or egret.Get("/",i18n.I18n(....), func (ctx *egret.Context){})

	egret.Get("/", func(ctx *egret.Context) {
		hi := ctx.GetFmt("translate")("hi", "maki") // hi is the key, 'maki' is the %s, the second parameter is optional
		language := ctx.Get("language") // language is the language key, example 'en-US'

		ctx.Write("From the language %s translated output: %s", language, hi)
	})

	egret.Listen(":8080")

}

```

### [For a working example, click here](https://github.com/kenorld/egret/tree/examples/middleware_internationalization_i18n)
