package api

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"go.uber.org/zap"
)

type responseSpeakerInfo struct {
	ID          int32    `jp:"id"`
	Name        string   `jp:"name"`
	IP          string   `jp:"ip"`
	MAC         string   `jp:"mac"`
	BitList     []string `jp:"bitList"`
	RateList    []int    `jp:"rateList"`
	Volume      int      `jp:"vol"`
	AbsoluteVol bool     `jp:"avol"`
	PowerState  int      `jp:"power"`

	Statistic speaker.Statistic `jp:"statisitc"`
}

type requestSpeakerInfo uint32

func apiSpeakerInfo(c *websocket.Conn, req Requester, log *zap.Logger) (any, error) {
	var sp requestSpeakerInfo
	err := req.Unmarshal(&sp)
	if err != nil {
		return nil, err
	}
	s := speaker.FindSpeakerByID(speaker.ID(sp))
	if s == nil {
		return nil, &Error{4, fmt.Errorf("speaker[%d] not exists", sp)}
	}
	power := int(s.PowerSate)
	if !s.PowerSave {
		power = -1
	}

	info := responseSpeakerInfo{
		ID:          int32(s.ID),
		Name:        s.Name,
		IP:          s.IP.String(),
		MAC:         s.MAC.String(),
		BitList:     s.BitsMask.Slice(),
		RateList:    s.RateMask.Slice(),
		Volume:      s.Volume,
		AbsoluteVol: s.AbsoluteVol,
		PowerState:  power,
		Statistic:   s.Statistic,
	}

	return info, nil
}
