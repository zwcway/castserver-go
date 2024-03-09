package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/dsp"
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

type requestLineEQ struct {
	ID        uint8   `jp:"id"`
	Seg       uint8   `jp:"seg"`  // 均衡器数量
	Frequency int     `jp:"freq"` // 频率
	Gain      float32 `jp:"gain"` // 增益
}

func apiLineSetEqualizer(c *websockets.WSConnection, req Requester, log lg.Logger) (any, error) {
	var p requestLineEQ
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}
	if !dsp.IsFrequencyValid(p.Frequency) {
		return nil, fmt.Errorf("frequency invalid")
	}
	if p.Seg > dsp.FEQ_MAX_SIZE {
		return nil, fmt.Errorf("seg invalid")
	}

	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		return nil, fmt.Errorf("line[%d] not exists", p.ID)
	}

	eq := nl.Equalizer()

	if len(eq.Filters) != int(p.Seg) {
		eq.Clear(p.Seg)
	}

	eq.Set(p.Frequency, float64(p.Gain), 0)

	if err = nl.SetEqualizer(eq); err != nil {
		return nil, err
	}

	on := nl.Input.EqualizerEle.IsOn()
	nl.Input.EqualizerEle.On()

	if !on {
		nl.Dispatch("line eq power", true)
	}

	return websockets.NewResponseEqualizer(nl), nil
}
