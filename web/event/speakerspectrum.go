package event

import (
	"context"

	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
	"go.uber.org/zap"
)

func speakerSpectrumRoutine(arg int, ctx context.Context, log *zap.Logger, ctrl chan int) {
	var ls stream.SpectrumElement

	sp := speaker.FindSpeakerByID(speaker.ID(arg))
	if sp == nil || sp.Spectrum == nil {
		return
	}
	ls = sp.Spectrum

	if ls.IsOn() {
		return
	}
	ls.On()

	defer func() {
		ls.Off()
		log.Info("stop speaker spectrum routine")
	}()
	log.Info("start speaker spectrum routine")

	spectrum(arg, ctx, log, ctrl, ls)
}
