package bus

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/utils"
)

var log lg.Logger
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

		if len(args) > 0 {
			log.Debug("dispatch", lg.String("event", e), lg.Int("count", int64(len(ll))), lg.Any("param", args[0]))
		} else {
			log.Debug("dispatch", lg.String("event", e), lg.Int("count", int64(len(ll))))
		}

		for _, hd := range ll {

			if hd.async {
				hd = hd.clone()
				hd.a = args
				queue <- hd
				continue
			}
			if hd.once > 1 {
				continue
			} else if hd.once == 1 {
				hd.once = 2
				removeHandler(hd)
			}

			e := hd.h(args...)
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
		if hd.once > 1 {
			continue
		} else if hd.once == 1 {
			hd.once = 2
		}

		hd.h(hd.a...)

		if hd.once > 0 {
			removeHandler(hd)
			hd.once = 3
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
	log = ctx.Logger("bus")
	go eventBusRoutine(ctx)
}
