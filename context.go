package tuan_frame

import (
	"net/http"
)

type HandlerFunc func(ctx *Context)

type Context struct {
	W http.ResponseWriter
	R *http.Request
}
