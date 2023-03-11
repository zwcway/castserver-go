package event

import (
	"context"

	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
	"go.uber.org/zap"
)

type notifySpectrum struct {
	LevelMeter [2]float32 `jp:"l"`
	Spectrum   []float32  `jp:"s"`
}

func lineSpectrumRoutine(arg int, ctx context.Context, log *zap.Logger, ctrl chan int) {
	var ls stream.SpectrumElement

	line := speaker.FindLineByID(speaker.LineID(arg))
	if line == nil {
		return
	}
	if ls = line.Spectrum; ls == nil {
		return
	}
	if ls.IsOn() {
		return
	}
	ls.On()

	defer func() {
		ls.Off()
		log.Info("stop line spectrum routine")
	}()
	log.Info("start line spectrum routine")

	spectrum(arg, ctx, log, ctrl, ls)
}
