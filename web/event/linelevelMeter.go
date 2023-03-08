package event

import (
	"context"
	"runtime"

	"github.com/zwcway/castserver-go/common/jsonpack"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/decoder/element"
	"github.com/zwcway/castserver-go/decoder/pipeline"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type notifyLineLevelMeter [2]float32

func lineLevelMeterRoutine(arg int, ctx context.Context, log *zap.Logger, ctrl chan int) {
	var ls *element.LineLevelMeter

	line := speaker.FindLineByID(speaker.LineID(arg))
	pl := pipeline.FromLine(line)
	if pl == nil {
		return
	}
	if ls = pl.EleLineLM(); ls == nil {
		return
	}
	if ls.State() {
		return
	}
	ls.On()

	log.Info("start line level meter routine")
	defer func() {
		ls.Off()
		log.Info("stop line level meter routine")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ctrl:
			return
		case <-ticker.C:
		}

		runtime.Gosched()

		if line.LevelMeter == 0 {
			continue
		}

		resp := notifyLineLevelMeter{
			float32(line.ID), float32(line.LevelMeter),
		}

		msg, err := jsonpack.Marshal(resp)
		if err == nil {
			websockets.Broadcast(websockets.Command_LINE, websockets.Event_Line_LevelMeter, 0, msg)
		}
	}
}
