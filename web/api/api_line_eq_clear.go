package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/dsp"
	log1 "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

type requestLineEQClear struct {
	ID  uint8 `jp:"id"`
	Seg uint8 `jp:"seg"`
}

func apiLineClearEqualizer(c *websockets.WSConnection, req Requester, log log1.Logger) (any, error) {
	var p requestLineEQClear
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}
	if p.Seg > dsp.FEQ_MAX_SIZE {
		return nil, fmt.Errorf("seg invalid")
	}

	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		return nil, fmt.Errorf("line[%d] not exists", p.ID)
	}

	nl.EqualizerEle.Off()
	nl.SetEqualizer(dsp.NewDataProcess(p.Seg))

	nl.Dispatch("line eq clean")
	nl.Dispatch("line eq power", false)

	return true, nil
}
