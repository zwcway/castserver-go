package api

import (
	"github.com/zwcway/castserver-go/common/audio"
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/sounds"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

type requestTestInfo struct {
	Line    *uint8  `jp:"line,omitempty"`
	Channel *uint8  `jp:"ch,omitempty"`
	Speaker *uint32 `jp:"sp,omitempty"`
}

func apiTestSound(c *websockets.WSConnection, req Requester, log lg.Logger) (ret any, err error) {
	var p requestTestInfo
	err = req.Unmarshal(&p)
	if err != nil {
		return
	}
	if p.Line != nil && p.Channel != nil {

		nl := speaker.FindLineByID(speaker.LineID(*p.Line))
		if nl == nil {
			err = &speaker.UnknownLineError{Line: *p.Line}
			return
		}

		ch := audio.Channel(*p.Channel)
		var chsound []byte

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
		nl.Input.PlayerEle.AddPCMWithChannel(ch, sounds.Format(), chsound)
	}

	if p.Speaker != nil {
		sp := speaker.FindSpeakerByID(speaker.SpeakerID(*p.Speaker))
		if sp != nil {
			sp.PlayerEle.AddPCMWithChannel(sp.SampleChannel(), sounds.Format(), sounds.Here())
		}
	}

	ret = true
	return
}
