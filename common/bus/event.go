package bus

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	log1 "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/utils"
)

var log log1.Logger
var list = make(map[string][]*HandlerData)
var queue = make(chan *HandlerData, 10)

func Register(e string, c Handler) (hd *HandlerData) {
	return RegisterObj(nil, e, c)
}

func RegisterObj(obj any, e string, c Handler) (hd *HandlerData) {
	if c == nil {
		panic(fmt.Errorf("callback can not nil for event(%s)", e))
	}

	if _, ok := list[e]; !ok {
		list[e] = make([]*HandlerData, 0)
	}

	hd = &HandlerData{
		e:   e,
		obj: obj,
		h:   c,
		hr:  reflect.ValueOf(c).Pointer(),
	}
	list[e] = append(list[e], hd)

	return hd
}

// 注销指定事件
//
// 忽略对象，按回调函数注销
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

// 注销绑定指定对象的所有事件
//
// 主要用于取消引用对象指针以GC
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
	return DispatchObj(nil, e, args...)
}

func DispatchObj(obj any, e string, args ...any) error {
	var (
		err   error
		count int = 0
	)

	if ll, ok := list[e]; ok {
		count = len(ll)

		for _, hd := range ll {
			if hd.obj != nil && hd.obj != obj {
				// 过滤指定对象的事件，不符合就跳过该回调
				continue
			}

			if hd.async {
				hd = hd.clone()
				hd.a = append([]any{obj}, args...)
				queue <- hd
				continue
			}
			if hd.once > 1 {
				continue
			} else if hd.once == 1 {
				hd.once = 2
				removeHandler(hd)
			}

			e := hd.h(obj, args...)
			if e != nil {
				if err != nil {
					err = errors.Wrap(e, "")
				} else {
					err = e
				}
			}
		}
	}

	lf := []log1.Field{log1.String("event", e), log1.Int("count", int64(count))}
	if len(args) > 0 {
		if s, ok := obj.(Eventer); ok {
			lf = append(lf, log1.String("from", s.Name()))
		} else if s, ok := obj.(fmt.Stringer); ok {
			lf = append(lf, log1.String("from", s.String()))
		}
	}
	log.Debug("dispatch", lf...)

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
		asyncCall(hd)
	}
}

func asyncCall(hd *HandlerData) {
	if hd.once > 1 {
		return
	} else if hd.once == 1 {
		hd.once = 2
	}

	hd.h(hd.a[0], hd.a[1:]...)
	hd.a = nil

	if hd.once > 0 {
		removeHandler(hd)
		hd.once = 3
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
