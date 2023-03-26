package bus

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/zwcway/castserver-go/common/utils"
)

var list = make(map[string][]*handlerData)
var queue = make(chan *handlerData, 10)
var lock sync.Mutex

func Register(e string, c Handler) (hd *handlerData) {
	if c == nil {
		panic(fmt.Errorf("callback can not nil for event(%s)", e))
	}

	lock.Lock()
	defer lock.Unlock()

	if _, ok := list[e]; !ok {
		list[e] = make([]*handlerData, 0)
	}

	hd = &handlerData{
		e:  e,
		h:  c,
		hr: reflect.ValueOf(c).Pointer(),
	}
	list[e] = append(list[e], hd)

	return hd
}

func Registers(c Handler, es ...string) {
	for i := 0; i < len(es); i++ {
		Register(es[i], c)
	}
}

func Dispatch(e string, args ...any) error {
	var err error

	if ll, ok := list[e]; ok {

		lock.Lock()
		defer lock.Unlock()

		for _, h := range ll {
			if h.async {
				h = h.clone()
				h.a = args
				queue <- h
				continue
			}
			e := h.h(args...)
			if e != nil {
				if err != nil {
					err = errors.Wrap(e, "")
				} else {
					err = e
				}
			}
		}
	}
	return err
}

// 异步执行
func eventBusRoutine(ctx utils.Context) {
	var hd *handlerData
	for {
		select {
		case <-ctx.Done():
			return
		case hd = <-queue:
		}
		if hd.once {
			lock.Lock()
		}

		hd.h(hd.a...)

		if hd.once {
			removeHandler(hd)
			lock.Unlock()
		}
	}
}

func removeHandler(h *handlerData) {
	if h == nil {
		return
	}

	for i, sp := range list[h.e] {
		if sp.hr == h.hr {
			s := list[h.e]
			utils.SliceQuickRemove(&s, i)
			list[h.e] = s
			return
		}
	}
}

func Init(ctx utils.Context) {
	go eventBusRoutine(ctx)
}
