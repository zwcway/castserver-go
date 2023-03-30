package bus

type Handler func(...any) error

type HandlerData struct {
	obj   any
	e     string
	h     Handler
	hr    uintptr
	a     []any
	once  uint8
	async bool
}

func (h *HandlerData) clone() *HandlerData {
	n := *h
	return &n
}

func (h *HandlerData) Once() *HandlerData {
	h.once = 1
	return h
}

func (h *HandlerData) ASync() *HandlerData {
	h.async = true
	return h
}

type Eventer interface {
	Dispatch(string, ...any) error
	Register(string, func(...any) error) *HandlerData
}
