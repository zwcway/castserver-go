package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/dsp"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestLineEQ struct {
	ID        uint8   `jp:"id"`
	Seg       uint8   `jp:"seg"`
	Frequency int     `jp:"freq"`
	Gain      float32 `jp:"gain"`
}

func apiLineSetEqualizer(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	var p requestLineEQ
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}
	if !dsp.IsFrequencyValid(p.Frequency) {
		return nil, fmt.Errorf("frequency invalid")
	}
	if p.Seg > dsp.FEQ_MAX_SIZE {
		return nil, fmt.Errorf("seg invalid")
	}

	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		return nil, fmt.Errorf("line[%d] not exists", p.ID)
	}

	eq := nl.Equalizer()

	if len(eq.FEQ) != int(p.Seg) {
		eq.Clear(p.Seg)
	}

	eq.AddFIR(p.Frequency, float64(p.Gain), 0)

	if err = nl.SetEqualizer(eq); err != nil {
		return nil, err
	}

	on := nl.EqualizerEle.IsOn()
	nl.EqualizerEle.On()

	if !on {
		bus.Dispatch("line eq power", nl, true)
	}

	return websockets.NewResponseEqualizer(nl), nil
}
