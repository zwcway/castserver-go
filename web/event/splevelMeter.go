package event

import (
	"context"
	"runtime"

	"github.com/zwcway/castserver-go/common/jsonpack"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type notifySpeakerLevelMeter [][2]float32

var levelMeterRuning = false

func speakerLevelMeterRoutine(arg int, ctx context.Context, log *zap.Logger, ctrl chan int) {
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
		case <-ctrl:
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
			websockets.Broadcast(websockets.Command_SPEAKER, websockets.Event_SP_LevelMeter, 0, msg)
		}
	}
}
