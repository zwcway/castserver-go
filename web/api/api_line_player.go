package api

import (
	log1 "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

func apiLinePlayer(c *websockets.WSConnection, req Requester, log log1.Logger) (any, error) {
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
