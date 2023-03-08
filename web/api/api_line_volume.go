package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/control"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

func apiLineVolume(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	var p requestVolume
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}

	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		return nil, fmt.Errorf("add new line faild")
	}

	nl.Volume = float64(p.Volume) / 100

	control.ControlLineVolume(nl, float64(p.Volume)/100, p.Mute)

	return true, nil
}
