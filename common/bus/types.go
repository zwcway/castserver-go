package bus

type Handler func(...any) error

type handlerList struct {
	e        string
	handlers []*handlerData
}
type handlerData struct {
	e    string
	h    Handler
	a    []any
	once bool
}

func (h *handlerData) Once() *handlerData {
	h.once = true
	return h
}
