package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

type requestSpeakerInfo uint32

func apiSpeakerInfo(c *websockets.WSConnection, req Requester, log lg.Logger) (any, error) {
	var p requestSpeakerInfo
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}
	sp := speaker.FindSpeakerByID(speaker.SpeakerID(p))
	if sp == nil {
		return nil, &Error{4, fmt.Errorf("speaker[%d] not exists", p)}
	}

	info := websockets.NewResponseSpeakerInfo(sp)

	return info, nil
}
