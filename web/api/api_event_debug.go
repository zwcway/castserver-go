package api

import (
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestEventDebug struct {
	Evt uint8 `jp:"evt"`
}

func apiEventDebug(c *websockets.WSConnection, req Requester, log *zap.Logger) (ret any, err error) {
	var p requestEventDebug

	err = req.Unmarshal(&p)
	if err != nil {
		return
	}
	if websockets.FindEvent(websockets.Command_SPEAKER, p.Evt) {
		sps := speaker.AllSpeakers()
		if len(sps) == 0 {
			ret = false
			return
		}
		websockets.BroadcastSpeakerEvent(sps[0], p.Evt)
		ret = true
		return
	} else if websockets.FindEvent(websockets.Command_LINE, p.Evt) {
		ls := speaker.LineList()
		if len(ls) == 0 {
			ret = false
			return
		}
		websockets.BroadcastLineEvent(ls[0], p.Evt)
		ret = true
		return
	}

	switch p.Evt {
	case websockets.Event_SP_LevelMeter:
	case websockets.Event_Line_LevelMeter:
	}

	ret = false
	return
}
