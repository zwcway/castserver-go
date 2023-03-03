//go:build !debug
// +build !debug

package api

import "github.com/valyala/fasthttp"

func ApiDispatchDevel(ctx *fasthttp.RequestCtx) bool {
	return false
}
