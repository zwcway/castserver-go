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
	Q         float32 `jp:"q"`
}

func apiLineSetEqualizer(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	var params requestLineEQ
	err := req.Unmarshal(&params)
	if err != nil {
		return nil, err
	}
	if !dsp.IsFrequencyValid(params.Frequency) {
		return nil, fmt.Errorf("frequency invalid")
	}

	nl := speaker.FindLineByID(speaker.LineID(params.ID))
	if nl == nil {
		return nil, fmt.Errorf("line[%d] not exists", params.ID)
	}

	for _, eq := range nl.Equalizer {
		if eq.Frequency == int(params.Frequency) {
			eq.Gain = params.Gain
			return true, nil
		}
	}

	nl.Equalizer = append(nl.Equalizer, dsp.FreqEqualizer{
		Frequency: params.Frequency,
		Gain:      params.Gain,
	})
	return true, nil
}
