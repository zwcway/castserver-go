package api

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/control"
	"go.uber.org/zap"
)

func apiLineVolume(c *websocket.Conn, req *ReqMessage, log *zap.Logger) (any, error) {
	var p requestVolume
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}

	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		return nil, fmt.Errorf("add new line faild")
	}

	control.ControlLineVolume(nl, int(p.Volume))

	return true, nil
}
