package bus

import (
	"fmt"

	"github.com/zwcway/castserver-go/utils"
)

var list = make(map[string][]*handlerData)
var queue = make(chan *handlerData, 10)

func Register(e string, c Handler) (hd *handlerData) {
	if c == nil {
		panic(fmt.Errorf("callback can not nil for event(%s)", e))
	}

	if _, ok := list[e]; !ok {
		list[e] = make([]*handlerData, 0)
	}

	hd = &handlerData{
		e: e,
		h: c,
	}
	list[e] = append(list[e], hd)

	return hd
}

func Trigger(e string, args ...any) {
	if ll, ok := list[e]; ok {
		for _, h := range ll {
			h.a = args
			queue <- h
		}
	}
}

// 异步执行
func routine(ctx utils.Context) {
	var hd *handlerData
	for {
		select {
		case <-ctx.Done():
			return
		case hd = <-queue:
		}

		hd.h(hd.a...)

		if hd.once {
			s := list[hd.e]
			utils.SliceQuickRemoveItem(&s, hd)
			list[hd.e] = s
		}
	}
}

func Init(ctx utils.Context) {
	go routine(ctx)
}
