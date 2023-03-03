package api

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"go.uber.org/zap"
)

type requestLineEdit struct {
	ID   uint8  `jp:"id"`
	Name string `jp:"name"`
}

func apiLineEdit(c *websocket.Conn, req Requester, log *zap.Logger) (ret any, err error) {
	var p requestLineEdit
	err = req.Unmarshal(&p)
	if err != nil {
		return
	}
	if len(p.Name) == 0 || len(p.Name) > 10 {
		err = fmt.Errorf("name invalid")
		return
	}
	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		err = fmt.Errorf("add new line faild")
		return
	}

	nl.Name = p.Name

	ret = true
	return
}
