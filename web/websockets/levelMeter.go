package websockets

import (
	"runtime"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/jsonpack"
	"github.com/zwcway/castserver-go/common/speaker"
)

type notifySpeakerLevelMeter [][2]float32

var levelMeterRuning = false
var levelMeterSignal = make(chan int, 1)
var ticker = time.NewTicker(200 * time.Millisecond)

func speakerLevelMeterRoutine(c *websocket.Conn) {
	if levelMeterRuning {
		return
	}
	levelMeterRuning = true

	defer func() {
		levelMeterRuning = false
		log.Info("stop speaker level meter routine")
	}()
	log.Info("start speaker level meter routine")

	for {
		select {
		case <-ctx.Done():
			return
		case <-levelMeterSignal:
			return
		case <-ticker.C:
		}

		runtime.Gosched()

		if speaker.CountSpeaker() == 0 {
			continue
		}

		resp := make(notifySpeakerLevelMeter, 0)
		speaker.All(func(s *speaker.Speaker) {
			resp = append(resp, [2]float32{float32(s.ID), float32(s.LevelMeter)})
		})
		msg, err := jsonpack.Marshal(resp)
		if err == nil {
			broadcast(Command_SPEAKER, Event_SP_LevelMeter, 0, msg)
		}
	}
}
