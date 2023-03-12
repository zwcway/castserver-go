package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/sounds"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestTestInfo struct {
	Line    *uint8  `jp:"line,omitempty"`
	Channel *uint8  `jp:"ch,omitempty"`
	Speaker *uint32 `jp:"sp,omitempty"`
}

func apiTestSound(c *websockets.WSConnection, req Requester, log *zap.Logger) (ret any, err error) {
	var p requestTestInfo
	err = req.Unmarshal(&p)
	if err != nil {
		return
	}
	if p.Line != nil && p.Channel != nil {

		nl := speaker.FindLineByID(speaker.LineID(*p.Line))
		if nl == nil {
			err = fmt.Errorf("add new line faild")
			return
		}

		ch := audio.Channel(*p.Channel)
		var chsound []float64

		switch ch {
		case audio.Channel_FRONT_LEFT:
			chsound = sounds.FrontLeft()
		case audio.Channel_FRONT_RIGHT:
			chsound = sounds.FrontRight()
		case audio.Channel_FRONT_CENTER:
			chsound = sounds.FrontCenter()
		case audio.Channel_SIDE_LEFT:
			chsound = sounds.SideLeft()
		case audio.Channel_SIDE_RIGHT:
			chsound = sounds.SideRight()
		case audio.Channel_BACK_LEFT:
			chsound = sounds.BackLeft()
		case audio.Channel_BACK_RIGHT:
			chsound = sounds.BackRight()
		case audio.Channel_BACK_CENTER:
			chsound = sounds.BackCenter()
		case audio.Channel_LOW_FREQUENCY:
			chsound = sounds.LowBass()
		default:
			return true, nil
		}

		nl.Player.Add(chsound)
	}

	if p.Speaker != nil {
		sp := speaker.FindSpeakerByID(speaker.ID(*p.Speaker))
		if sp != nil {
			sp.Player.Add(sounds.Here())
		}
	}

	ret = true
	return
}
