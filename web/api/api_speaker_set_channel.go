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
	Name    string `jp:"name"`
	Channel int8   `jp:"ch"`
}

func apiSpeakerEdit(c *websocket.Conn, req Requester, log *zap.Logger) (any, error) {
	var p requestSpeakerSetChannel
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}
	s := speaker.FindSpeakerByID(speaker.ID(p.ID))
	if s == nil {
		return nil, &Error{4, fmt.Errorf("speaker[%d] not exists", p.ID)}
	}
	if len(p.Name) > 0 {

	} else if p.Channel > 0 {
		ch := audio.Channel(p.Channel)
		s.ChangeChannel(ch)
	} else if p.Channel == -1 {
		s.ChangeChannel(0)
	}

	return true, nil
}
