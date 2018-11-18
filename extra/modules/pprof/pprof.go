package pprof

import (
	"net/http/pprof"
	"strings"

	"github.com/kataras/egret"
)

// New returns the pprof (profile, debug usage) Handler/ middleware
// NOTE: Route MUST have the last named parameter wildcard named '*action'
// Usage:
// egret.Get("debug/pprof/*action", pprof.New())
func New() egret.HandlerFunc {
	indexHandler := egret.ToHandler(pprof.Index)
	cmdlineHandler := egret.ToHandler(pprof.Cmdline)
	profileHandler := egret.ToHandler(pprof.Profile)
	symbolHandler := egret.ToHandler(pprof.Symbol)
	goroutineHandler := egret.ToHandler(pprof.Handler("goroutine"))
	heapHandler := egret.ToHandler(pprof.Handler("heap"))
	threadcreateHandler := egret.ToHandler(pprof.Handler("threadcreate"))
	debugBlockHandler := egret.ToHandler(pprof.Handler("block"))

	return egret.HandlerFunc(func(ctx *egret.Context) {
		ctx.SetContentType("text/html; charset=" + ctx.Framework().Config.Charset)

		action := ctx.Param("action")
		if len(action) > 1 {
			if strings.Contains(action, "cmdline") {
				cmdlineHandler.Serve((ctx))
			} else if strings.Contains(action, "profile") {
				profileHandler.Serve(ctx)
			} else if strings.Contains(action, "symbol") {
				symbolHandler.Serve(ctx)
			} else if strings.Contains(action, "goroutine") {
				goroutineHandler.Serve(ctx)
			} else if strings.Contains(action, "heap") {
				heapHandler.Serve(ctx)
			} else if strings.Contains(action, "threadcreate") {
				threadcreateHandler.Serve(ctx)
			} else if strings.Contains(action, "debug/block") {
				debugBlockHandler.Serve(ctx)
			}
		} else {
			indexHandler.Serve(ctx)
		}
	})
}
