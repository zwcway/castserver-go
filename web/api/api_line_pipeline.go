package api

import (
	"fmt"

	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/pipeline"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

type responseLinePipeLineSpend struct {
	Name string `jp:"name"`
	Cost int    `jp:"ms"`
}
type responseLinePipeLine struct {
	Dlna   bool `jp:"dlna"`
	Spends []responseLinePipeLineSpend
}

func apiLinePipeLineInfo(c *websockets.WSConnection, req Requester, log lg.Logger) (any, error) {
	var p requestLineInfo
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}

	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		return nil, fmt.Errorf("line %d not exists", p.ID)
	}

	pl, ok := nl.Input.PipeLine.(*pipeline.PipeLine)
	if !ok {
		return nil, fmt.Errorf("no found")
	}

	resp := responseLinePipeLine{
		Dlna:   false,
		Spends: make([]responseLinePipeLineSpend, 0),
	}
	for _, s := range pl.Streamers() {
		resp.Spends = append(resp.Spends, responseLinePipeLineSpend{
			Name: s.Name(),
			Cost: int(s.Cost().Milliseconds()),
		})
	}

	return resp, nil
}
