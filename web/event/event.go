package event

import (
	"context"
	"sync"
	"time"

	"github.com/zwcway/castserver-go/utils"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type cbHandler func(arg int, ctx context.Context, log *zap.Logger, ctrl chan int)

var ticker = time.NewTicker(50 * time.Millisecond)
var handler = websockets.EventHandler{On: start, Off: stop}

// 事件回调列表
var EventHandlerMap = map[uint8]websockets.EventHandler{
	websockets.Event_Line_Spectrum:   handler,
	websockets.Event_Line_LevelMeter: handler,
	websockets.Event_SP_LevelMeter:   handler,
}

var services = []eventService{}

type eventService struct {
	signal chan int
	arg    int
	evt    uint8
	on     cbHandler
}

var locker sync.Mutex

func start(evt uint8, arg int, ctx context.Context, log *zap.Logger) {
	locker.Lock()
	defer locker.Unlock()

	for _, es := range services {
		if es.evt == evt && es.arg == arg {
			return
		}
	}
	es := eventService{
		arg: arg,
		evt: evt,
	}
	switch evt {
	case websockets.Event_Line_Spectrum, websockets.Event_Line_LevelMeter:
		es.on = lineSpectrumRoutine
	case websockets.Event_SP_LevelMeter:
		es.on = speakerSpectrumRoutine
	default:
		return
	}
	es.signal = make(chan int, 2)

	services = append(services, es)

	// 启动事件推送
	go es.on(arg, ctx, log, es.signal)
}

func stop(evt uint8, arg int) {
	for i, es := range services {
		if es.evt == evt {
			es.signal <- 1
			services = utils.SliceRemove(services, i)
			return
		}
	}
}
