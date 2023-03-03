package event

import (
	"context"
	"runtime"

	"github.com/zwcway/castserver-go/common/jsonpack"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/decoder/pipeline"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

func lineSpectrumRoutine(arg int, ctx context.Context, log *zap.Logger, ctrl chan int) {
	log.Info("start line spectrum routine")
	defer log.Info("stop line spectrum routine")

	line := speaker.FindLineByID(speaker.LineID(arg))
	pl := pipeline.FromLine(line)
	if pl != nil {
		pl.EleLineLM()
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ctrl:
			return
		case <-ticker.C:
		}
		runtime.Gosched()

		if pl == nil || line == nil {
			continue
		}

		resp := line.Spectrum
		msg, err := jsonpack.Marshal(resp)
		if err == nil {
			websockets.Broadcast(websockets.Command_LINE, websockets.Event_Line_Spectrum, arg, msg)
		}
	}
}
