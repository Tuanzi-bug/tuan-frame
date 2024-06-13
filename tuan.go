package tuan_frame

import (
	"log"
	"net/http"
)

const (
	ANY = "ANY"
)

type Engine struct {
	router
}

func New() *Engine {
	return &Engine{
		router{},
	}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	groups := e.router.groups
	for _, g := range groups {
		for name, methodHandle := range g.handlerMap {
			url := g.groupName + name
			if r.RequestURI == url {
				ctx := &Context{
					W: w,
					R: r,
				}
				if _, ok := methodHandle[ANY]; ok {
					methodHandle[ANY](ctx)
				}
				method := r.Method
				handle, ok := methodHandle[method]
				if ok {
					handle(ctx)
					return
				}
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			} else {
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}
	}
}

func (e *Engine) Run() {
	http.Handle("/", e)
	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}
