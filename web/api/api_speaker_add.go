package api

import (
	"net"
	"net/netip"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/detector"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestSpeakerCreate struct {
	Ver      uint8
	ID       uint32
	IP       string
	MAC      string
	DataPort uint16
	BitsMask []uint8
	RateMask []uint8
	AVol     bool
}

func apiSpeakerCreate(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	p := requestSpeakerCreate{}
	if err := req.Unmarshal(&p); err != nil {
		return nil, err
	}

	mac, err := net.ParseMAC(p.MAC)
	if err != nil {
		return nil, err
	}
	rm, err := audio.NewAudioRateMask(p.RateMask)
	if err != nil {
		return nil, err
	}
	bm, err := audio.NewAudioBitsMask(p.BitsMask)
	if err != nil {
		return nil, err
	}
	res := &detector.SpeakerResponse{
		Ver:        uint8(p.Ver),
		ID:         speaker.ID(p.ID),
		Connected:  false,
		Addr:       netip.MustParseAddr(p.IP),
		MAC:        mac,
		RateMask:   rm,
		BitsMask:   bm,
		DataPort:   p.DataPort,
		AbsolueVol: p.AVol,
		PowerSave:  true,
	}
	err = detector.CheckSpeaker(res)
	if err != nil {
		return nil, err
	}
	return true, nil
}
