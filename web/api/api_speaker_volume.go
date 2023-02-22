package api

import (
	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/control"
	"go.uber.org/zap"
)

type requestVolume struct {
	ID     int32  `jp:"id"`
	Volume uint32 `jp:"vol"`
}

func apiSpeakerVolume(c *websocket.Conn, req *ReqMessage, log *zap.Logger) (any, error) {
	var sp requestVolume
	err := req.Unmarshal(&sp)
	if err != nil {
		return nil, err
	}
	s := speaker.FindSpeakerByID(speaker.ID(sp.ID))
	if s == nil {
		return nil, nil
	}

	control.ControlSpeakerVolume(s, int(sp.Volume))
	return nil, nil
}
