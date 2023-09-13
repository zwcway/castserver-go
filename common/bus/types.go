package bus

type Handler = func(
	o any, // 触发事件的对象，注：同一事件名称允许不同对象任意调用
	a ...any, // 事件参数列表
) error

type HandlerData struct {
	e     string  // 事件名称，是唯一主键
	obj   any     // 仅用于事件按对象过滤
	h     Handler // 事件回调函数
	hr    uintptr // 回调函数指针，用于按回调函数查找
	a     []any   // 回调参数列表，用于异步回调
	once  uint8   // 是否调用一次即销毁
	async bool    // 是否异步回调
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
	Name() string
	Dispatch(string, ...any) error
	Register(string, Handler) *HandlerData
}
