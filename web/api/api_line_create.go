package api

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/receiver"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestLineCreate struct {
	Name string `jp:"name"`
}

func apiLineCreate(c *websocket.Conn, req Requester, log *zap.Logger) (any, error) {
	var params requestLineCreate
	err := req.Unmarshal(&params)
	if err != nil {
		return nil, err
	}
	if len(params.Name) == 0 || len(params.Name) > 10 {
		return nil, fmt.Errorf("name invalid")
	}

	nl := speaker.AddLine(params.Name)
	if nl == nil {
		return nil, fmt.Errorf("add new line faild")
	}

	line := websockets.NewResponseLineInfo(nl)

	receiver.AddDLNA(nl)
	websockets.BroadcastLineEvent(nl, websockets.Event_Line_Created)

	return line, nil
}
