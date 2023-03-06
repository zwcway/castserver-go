package api

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestLineInfo struct {
	ID uint8 `jp:"id"`
}

func apiLineInfo(c *websocket.Conn, req Requester, log *zap.Logger) (any, error) {
	var params requestLineInfo
	err := req.Unmarshal(&params)
	if err != nil {
		return nil, err
	}

	nl := speaker.FindLineByID(speaker.LineID(params.ID))
	if nl == nil {
		return nil, fmt.Errorf("add new line faild")
	}

	line := websockets.NewResponseLineInfo(nl)
	return line, nil
}
