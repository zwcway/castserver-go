package api

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"go.uber.org/zap"
)

type requestSpeakerSetChannel struct {
	ID      uint32 `jp:"id"`
	Channel uint8  `jp:"ch"`
}

func apiSpeakerSetChannel(c *websocket.Conn, req *ReqMessage, log *zap.Logger) (any, error) {
	var p requestSpeakerSetChannel
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}
	ch := audio.Channel(p.Channel)
	s := speaker.FindSpeakerByID(speaker.ID(p.ID))
	if s == nil {
		return nil, &Error{4, fmt.Errorf("speaker[%d] not exists", p.ID)}
	}
	s.ChangeChannel(ch)

	return true, nil
}
