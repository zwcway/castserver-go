package api

import (
	"fmt"
	"unicode/utf8"

	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/receiver"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestLineEdit struct {
	ID   uint8  `jp:"id"`
	Name string `jp:"name"`
}

func apiLineEdit(c *websockets.WSConnection, req Requester, log *zap.Logger) (ret any, err error) {
	var p requestLineEdit
	err = req.Unmarshal(&p)
	if err != nil {
		return
	}
	if len(p.Name) == 0 || utf8.RuneCountInString(p.Name) > 10 {
		err = fmt.Errorf("name invalid")
		return
	}
	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		err = fmt.Errorf("add new line faild")
		return
	}

	nl.Name = p.Name
	receiver.EditDLNA(nl)

	ret = true
	return
}
