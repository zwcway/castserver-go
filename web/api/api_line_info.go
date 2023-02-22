package api

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"go.uber.org/zap"
)

type requestLineInfo struct {
	ID uint8 `jp:"id"`
}

type responseSource struct {
	Rate     int    `jp:"rate"`
	Bits     string `jp:"bits"`
	Channels int    `jp:"channels"`
}
type responseLineInfo struct {
	ID     uint8  `jp:"id"`
	Name   string `jp:"name"`
	Volume int    `jp:"vol"`

	Speakers []responseSpeakerList `jp:"speakers,omitempty"`
	Input    *responseSource       `jp:"source,omitempty"`
}

func apiLineInfo(c *websocket.Conn, req *ReqMessage, log *zap.Logger) (any, error) {
	var params requestLineInfo
	err := req.Unmarshal(&params)
	if err != nil {
		return nil, err
	}

	nl := speaker.FindLineByID(speaker.LineID(params.ID))
	if nl == nil {
		return nil, fmt.Errorf("add new line faild")
	}

	line := responseLineInfo{
		ID:     uint8(nl.ID),
		Name:   nl.Name,
		Volume: nl.Volume,

		Speakers: make([]responseSpeakerList, speaker.CountLineSpeaker(nl.ID)),
		Input: &responseSource{
			Rate:     nl.Input.SampleRate.ToInt(),
			Bits:     nl.Input.SampleBits.Name(),
			Channels: nl.Input.Layout.Count,
		},
	}

	for i, s := range speaker.FindSpeakersByLine(nl.ID) {
		line.Speakers[i] = responseSpeakerList{
			ID:      int32(s.ID),
			Name:    s.Name,
			IP:      s.IP.String(),
			MAC:     s.MAC.String(),
			Channel: int(s.Channel),
			Volume:  s.Volume,
		}
	}

	return line, nil
}
