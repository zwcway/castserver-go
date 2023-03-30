package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

type requestLineCreate struct {
	Name string `jp:"name"`
}

func apiLineCreate(c *websockets.WSConnection, req Requester, log lg.Logger) (any, error) {
	var params requestLineCreate
	err := req.Unmarshal(&params)
	if err != nil {
		return nil, err
	}
	if len(params.Name) == 0 || len(params.Name) > 10 {
		return nil, fmt.Errorf("name invalid")
	}
	if speaker.CountLine() >= int(speaker.LineID_MAX) {
		return nil, fmt.Errorf("more than 255")
	}

	nl := speaker.NewLine(params.Name)
	if nl == nil {
		return nil, fmt.Errorf("add new line faild")
	}

	line := websockets.NewResponseLineInfo(nl)

	return line, nil
}
