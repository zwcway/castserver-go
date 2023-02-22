package api

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type reqSubscribe struct {
	Command uint8   `jp:"evt"`
	Action  bool    `jp:"act"`
	Event   []uint8 `jp:"sub,omitempty"`
	Arg     int     `jp:"arg"`
}

var SubscribeFunction func(c *websocket.Conn, evt int)
var UnsubscribeFunction func(c *websocket.Conn, evt int)

func apiSubscribe(c *websocket.Conn, req *ReqMessage, log *zap.Logger) (any, error) {
	params := reqSubscribe{}

	err := req.Unmarshal(&params)
	if err != nil {
		return nil, err
	}
	if websockets.Command_MIN >= params.Command || params.Command >= uint8(websockets.Command_MAX) {
		return nil, fmt.Errorf("command invalid")
	}

	if params.Action {
		websockets.Subscribe(c, params.Command, params.Event, params.Arg)
	} else {
		websockets.Unsubscribe(c, params.Command, params.Event, params.Arg)
	}
	return true, nil
}
