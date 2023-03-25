package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/dsp"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

func apiLineClearEqualizer(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	var params requestLineInfo
	err := req.Unmarshal(&params)
	if err != nil {
		return nil, err
	}

	nl := speaker.FindLineByID(speaker.LineID(params.ID))
	if nl == nil {
		return nil, fmt.Errorf("line[%d] not exists", params.ID)
	}

	nl.Equalizer.SetEqualizer([]dsp.Equalizer{})
	nl.Equalizer.Off()

	bus.Trigger("line equalizer clean", nl)
	bus.Trigger("line equalizer power", nl, false)

	return true, nil
}
