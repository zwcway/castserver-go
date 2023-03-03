package api

import (
	"time"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/speaker"
	"go.uber.org/zap"
)

type responseSpeakerList struct {
	ID          int32    `jp:"id"`
	Name        string   `jp:"name"`
	IP          string   `jp:"ip"`
	MAC         string   `jp:"mac"`
	Channel     int      `jp:"ch"`
	BitList     []string `jp:"bitList,omitempty"`
	RateList    []int    `jp:"rateList,omitempty"`
	Volume      int      `jp:"vol,omitempty"`
	AbsoluteVol bool     `jp:"avol,omitempty"`
	PowerState  int      `jp:"power,omitempty"`
	ConnectTime int      `jp:"cTime,omitempty"`
}

func apiSpeakerList(c *websocket.Conn, req Requester, log *zap.Logger) (any, error) {
	list := []responseSpeakerList{}

	speaker.All(func(s *speaker.Speaker) {
		power := int(s.PowerSate)
		if !s.PowerSave {
			power = -1
		}
		ct := 0
		if !s.ConnTime.IsZero() {
			ct = int(time.Since(s.ConnTime) / time.Second)
		}
		list = append(list, responseSpeakerList{
			ID:          int32(s.ID),
			Name:        s.Name,
			IP:          s.IP.String(),
			MAC:         s.MAC.String(),
			BitList:     s.BitsMask.Slice(),
			RateList:    s.RateMask.Slice(),
			Volume:      s.Volume,
			AbsoluteVol: s.AbsoluteVol,
			PowerState:  power,
			ConnectTime: ct,
		})
	})

	return list, nil
}
