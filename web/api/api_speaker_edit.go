package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/control"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestSpeakerSetChannel struct {
	ID      uint32 `jp:"id"`
	Name    string `jp:"name,omitempty"`
	Channel int8   `jp:"ch,omitempty"`
	Volume  *uint8 `jp:"vol,omitempty"`
	Mute    *bool  `jp:"mute,omitempty"`
}

func apiSpeakerEdit(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	var p requestSpeakerSetChannel
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}
	sp := speaker.FindSpeakerByID(speaker.ID(p.ID))
	if sp == nil {
		return nil, &Error{4, fmt.Errorf("speaker[%d] not exists", p.ID)}
	}
	if len(p.Name) > 0 {
		sp.Name = p.Name
	}
	if p.Channel > 0 {
		ch := audio.Channel(p.Channel)
		sp.ChangeChannel(ch)
	} else if p.Channel == -1 {
		sp.ChangeChannel(0)
	}
	if p.Volume != nil || p.Mute != nil {
		control.ControlSpeakerVolume(sp, float64(*p.Volume)/100, *p.Mute)
	}

	return true, nil
}
