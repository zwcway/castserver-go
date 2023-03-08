package api

import (
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/control"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestVolume struct {
	ID     int32  `jp:"id"`
	Volume uint32 `jp:"vol,omitempty"`
	Mute   bool   `jp:"mute,omitempty"`
}

func apiSpeakerVolume(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	var p requestVolume
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}
	sp := speaker.FindSpeakerByID(speaker.ID(p.ID))
	if sp == nil {
		return nil, nil
	}

	sp.Volume = int(p.Volume)
	control.ControlSpeakerVolume(sp)
	return true, nil
}
