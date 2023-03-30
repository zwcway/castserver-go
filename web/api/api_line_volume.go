package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

func apiLineVolume(c *websockets.WSConnection, req Requester, log lg.Logger) (any, error) {
	var p requestVolume
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}

	line := speaker.FindLineByID(speaker.LineID(p.ID))
	if line == nil {
		return nil, fmt.Errorf("line %d not exists", p.ID)
	}

	if p.Volume != nil {
		line.SetVolume(*p.Volume, line.Mute)
	} else if p.Mute != nil {
		line.SetVolume(line.Volume, *p.Mute)
	}

	return true, nil
}
