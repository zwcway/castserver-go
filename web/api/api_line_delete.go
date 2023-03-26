package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestLineDelete struct {
	ID   uint8 `jp:"id"`
	Move uint8 `jp:"moveTo,omitempty"`
}

func apiLineDelete(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	var params requestLineDelete
	err := req.Unmarshal(&params)
	if err != nil {
		return nil, err
	}

	nl := speaker.FindLineByID(speaker.LineID(params.ID))
	if nl == nil {
		return nil, fmt.Errorf("line[%d] not exists", params.ID)
	}

	if params.Move == 0 {
		params.Move = speaker.DefaultLineID
	}

	err = speaker.DelLine(nl.ID, speaker.LineID(params.Move))
	if err != nil {
		return nil, err
	}

	return nl.ID, nil
}
