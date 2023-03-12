package event

import (
	"context"
	"runtime"

	"github.com/zwcway/castserver-go/common/jsonpack"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

func spectrum(arg int, ctx context.Context, log *zap.Logger, ctrl chan int, ls stream.SpectrumElement) {
	for {
		select {
		case <-ctx.Done(): // 全局退出
			return
		case <-ctrl: // routine 退出
			return
		case <-ticker.C: // 定时器
		}

		runtime.Gosched()
		st := ls.Spectrum()

		if len(st) == 0 && ls.LevelMeter() == 0 {
			continue
		}

		resp := notifySpectrum{
			LevelMeter: [2]float32{float32(arg), float32(ls.LevelMeter())},
			Spectrum:   make([]float32, len(st)),
		}

		for i := 0; i < len(st); i++ {
			resp.Spectrum[i] = float32(st[i])
		}

		msg, err := jsonpack.Marshal(resp)
		if err == nil {
			websockets.Broadcast(websockets.Event_Line_Spectrum, 0, arg, msg)
		}
	}
}
