package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

type requestLineEQSwitch struct {
	ID     uint8 `jp:"id"`
	Enable bool  `jp:"enable"`
}

func apiLineSetEqualizerEnable(c *websockets.WSConnection, req Requester, log lg.Logger) (any, error) {
	var p requestLineEQSwitch
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}

	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		return nil, fmt.Errorf("line[%d] not exists", p.ID)
	}

	if p.Enable {
		nl.EqualizerEle.On()
	} else {
		nl.EqualizerEle.Off()
	}

	nl.Dispatch("line eq power", p.Enable)

	return true, nil
}
