package bus

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/utils"
)

var log lg.Logger
var list = make(map[string][]*HandlerData)
var queue = make(chan *HandlerData, 10)

func Register(e string, c Handler) (hd *HandlerData) {
	if c == nil {
		panic(fmt.Errorf("callback can not nil for event(%s)", e))
	}

	if _, ok := list[e]; !ok {
		list[e] = make([]*HandlerData, 0)
	}

	hd = &HandlerData{
		e:  e,
		h:  c,
		hr: reflect.ValueOf(c).Pointer(),
	}
	list[e] = append(list[e], hd)

	return hd
}

func RegisterObj(obj any, e string, c Handler) (hd *HandlerData) {
	hd = Register(e, c)
	hd.obj = obj
	return
}

func Unregister(e string, c Handler) {
	if ll, ok := list[e]; ok {
		cr := reflect.ValueOf(c).Pointer()
		for _, hd := range ll {
			if hd.hr == cr {
				removeHandler(hd)
				return
			}
		}
	}
}

func UnregisterObj(obj any) {
	if obj == nil {
		return
	}
	for e, ll := range list {
		nl := []*HandlerData{}
		for _, hd := range ll {
			if hd.obj != obj {
				nl = append(nl, hd)
			}
		}
		if len(nl) == 0 {
			delete(list, e)
		} else if len(nl) != len(ll) {
			list[e] = nl
		}
	}
}

func Dispatch(e string, args ...any) error {
	var (
		err   error
		count int = 0
	)

	if ll, ok := list[e]; ok {
		count = len(ll)

		for _, hd := range ll {
			if hd.obj != nil && len(args) > 0 && hd.obj != args[0] {
				continue
			}

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

	if len(args) > 0 {
		if s, ok := args[0].(fmt.Stringer); ok {
			log.Debug("dispatch", lg.String("event", e), lg.Int("count", int64(count)), lg.String("param", s.String()))
			return err
		}
	}
	log.Debug("dispatch", lg.String("event", e), lg.Int("count", int64(count)))

	return err
}

// 异步执行
func eventBusRoutine(ctx utils.Context) {
	var hd *HandlerData
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

func removeHandler(h *HandlerData) {
	if h == nil {
		return
	}

	for i, sp := range list[h.e] {
		if sp.hr == h.hr {
			s := list[h.e]
			utils.SliceQuickRemove(&s, i)
			if len(s) == 0 {
				delete(list, h.e)
			} else {
				list[h.e] = s
			}
			return
		}
	}
}

func Init(ctx utils.Context) {
	log = ctx.Logger("bus")
	go eventBusRoutine(ctx)
}
