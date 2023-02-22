package api

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"go.uber.org/zap"
)

type requestLineDelete struct {
	ID   uint8 `jp:"id"`
	Move uint8 `jp:"moveTo"`
}

func apiLineDelete(c *websocket.Conn, req *ReqMessage, log *zap.Logger) (any, error) {
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

	return nl.ID, nil
}
