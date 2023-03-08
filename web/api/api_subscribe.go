package api

import (
	"github.com/zwcway/castserver-go/web/event"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type reqSubscribe struct {
	Action bool    `jp:"act"`
	Event  []uint8 `jp:"evt"`
	SubEvt uint8   `jp:"sub,omitempty"`
	Arg    int     `jp:"arg,omitempty"`
}

var SubscribeFunction func(c *websockets.WSConnection, evt int)
var UnsubscribeFunction func(c *websockets.WSConnection, evt int)

func apiSubscribe(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	if c == nil {
		return nil, nil
	}
	params := reqSubscribe{}

	err := req.Unmarshal(&params)
	if err != nil {
		return nil, err
	}

	websockets.SetEventHandler(event.EventHandlerMap)

	if params.Action {
		websockets.Subscribe(c, params.Event, params.SubEvt, params.Arg)
	} else {
		websockets.Unsubscribe(c, params.Event, params.SubEvt, params.Arg)
	}

	return true, nil
}
