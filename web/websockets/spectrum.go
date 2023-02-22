package websockets

import (
	"runtime"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/jsonpack"
	"github.com/zwcway/castserver-go/common/speaker"
)

var lineSpectrumRuning = false
var lineSpectrumSignal = make(chan int, 1)

func lineSpectrumRoutine(c *websocket.Conn, lineId int) {
	if lineSpectrumRuning {
		return
	}
	lineSpectrumRuning = true
	defer func() {
		lineSpectrumRuning = false
		log.Info("stop line spectrum routine")
	}()
	log.Info("start line spectrum routine")

	for {
		select {
		case <-ctx.Done():
			return
		case <-lineSpectrumSignal:
			return
		case <-ticker.C:
		}

		runtime.Gosched()

		line := speaker.FindLineByID(speaker.LineID(lineId))
		if line == nil {
			return
		}

		resp := line.Spectrum
		msg, err := jsonpack.Marshal(resp)
		if err == nil {
			broadcast(Command_LINE, Event_Line_Spectrum, lineId, msg)
		}
	}
}
