package api

import (
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestVolume struct {
	ID     int32  `jp:"id"`
	Volume *uint8 `jp:"vol,omitempty"`
	Mute   *bool  `jp:"mute,omitempty"`
}

func apiSpeakerVolume(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	var p requestVolume
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}
	sp := speaker.FindSpeakerByID(speaker.SpeakerID(p.ID))
	if sp == nil {
		return nil, nil
	}

	if p.Volume != nil {
		sp.SetVolume(*p.Volume, sp.Mute)
	} else if p.Mute != nil {
		sp.SetVolume(sp.Volume, *p.Mute)
	}

	return true, nil
}
