package api

import (
	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

type requestLineInfo struct {
	ID uint8 `jp:"id"`
}

func apiLineInfo(c *websockets.WSConnection, req Requester, log lg.Logger) (any, error) {
	var params requestLineInfo
	err := req.Unmarshal(&params)
	if err != nil {
		return nil, err
	}

	nl := speaker.FindLineByID(speaker.LineID(params.ID))
	if nl == nil {
		return nil, &speaker.UnknownLineError{Line: params.ID}
	}

	line := websockets.NewResponseLineInfo(nl)
	return line, nil
}
