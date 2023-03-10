package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/dsp"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestLineEQ struct {
	ID        uint8   `jp:"id"`
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

	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		return nil, fmt.Errorf("line[%d] not exists", p.ID)
	}

	eqs := nl.Equalizer.Equalizer()

	for i := 0; i < len(eqs); i++ {
		if eqs[i].Frequency == int(p.Frequency) {
			eqs[i].Gain = float64(p.Gain)
			return true, nil
		}
	}

	nl.Equalizer.Add(p.Frequency, float64(p.Gain), 0)

	nl.Equalizer.On()

	return true, nil
}
