package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/receiver"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestLineDelete struct {
	ID   uint8 `jp:"id"`
	Move uint8 `jp:"moveTo"`
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

	err = speaker.DelLine(nl.ID, speaker.LineID(params.Move))
	if err != nil {
		return nil, err
	}

	receiver.DelDLNA(nl)
	websockets.BroadcastLineEvent(nl, websockets.Event_Line_Deleted)

	return nl.ID, nil
}
