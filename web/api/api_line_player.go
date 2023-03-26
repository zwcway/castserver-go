package api

import (
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

func apiLinePlayer(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	var params requestLineInfo
	err := req.Unmarshal(&params)
	if err != nil {
		return nil, err
	}

	nl := speaker.FindLineByID(speaker.LineID(params.ID))
	if nl == nil {
		return nil, &speaker.UnknownLineError{Line: params.ID}
	}

	line := websockets.NewResponseLineSource(nl)
	return line, nil
}
