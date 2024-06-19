package tuan_frame

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
