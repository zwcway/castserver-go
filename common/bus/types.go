package bus

type Handler func(...any) error

type handlerData struct {
	e     string
	h     Handler
	hr    uintptr
	a     []any
	once  uint8
	async bool
}

func (h *handlerData) clone() *handlerData {
	n := *h
	return &n
}

func (h *handlerData) Once() *handlerData {
	h.once = 1
	return h
}

func (h *handlerData) ASync() *handlerData {
	h.async = true
	return h
}
