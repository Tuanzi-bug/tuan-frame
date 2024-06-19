package tuan_frame

import (
	"net/http"
)

type HandlerFunc func(ctx *Context)

type Context struct {
	Req        *http.Request
	Resp       http.ResponseWriter
	PathParams map[string]string
}
