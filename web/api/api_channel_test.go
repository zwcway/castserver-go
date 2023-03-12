package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestChannelInfo struct {
	Line    uint8 `jp"line"`
	Channel uint8 `jp:"ch"`
}

func apiChannelPlayTest(c *websockets.WSConnection, req Requester, log *zap.Logger) (ret any, err error) {
	var p requestChannelInfo
	err = req.Unmarshal(&p)
	if err != nil {
		return
	}

	nl := speaker.FindLineByID(speaker.LineID(p.Line))
	if nl == nil {
		err = fmt.Errorf("add new line faild")
		return
	}

	for _, sp := range nl.SpeakersByChannel(audio.Channel(p.Channel)) {
		sp.Mixer.Add()
	}

	ret = true
	return
}
