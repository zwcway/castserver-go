package api

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/decoder/pipeline"
	"go.uber.org/zap"
)

type responseLinePipeLineSpend struct {
	Name string `jp:"name"`
	Cost int    `jp:"ms"`
}
type responseLinePipeLine struct {
	Dlna   bool `jp:"dlna"`
	Spends []responseLinePipeLineSpend
}

func apiLinePipeLineInfo(c *websocket.Conn, req Requester, log *zap.Logger) (any, error) {
	var p requestLineInfo
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}

	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		return nil, fmt.Errorf("add new line faild")
	}

	pl := pipeline.FromLine(nl)
	if pl == nil {
		return nil, fmt.Errorf("no found")
	}

	resp := responseLinePipeLine{
		Dlna:   false,
		Spends: make([]responseLinePipeLineSpend, 0),
	}
	for _, s := range pl.Streamers() {
		resp.Spends = append(resp.Spends, responseLinePipeLineSpend{
			Name: s.Name(),
			Cost: s.Cost(),
		})
	}

	return resp, nil
}
