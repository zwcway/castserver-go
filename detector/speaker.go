package detector

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/control"
	"github.com/zwcway/castserver-go/pusher"
	"github.com/zwcway/castserver-go/utils"

	"go.uber.org/zap"
)

func initSpeaker(sp *speaker.Speaker, res *SpeakerResponse) {
	sp.Name = res.MAC.String()
	sp.Config.RateMask = res.RateMask
	sp.Config.BitsMask = res.BitsMask
	sp.Config.Dport = res.DataPort
	sp.Config.MAC = res.MAC
	sp.Config.SetIP(res.Addr)
	sp.Config.AbsoluteVol = res.AbsoluteVol
	sp.Config.PowerSave = res.PowerSave
	sp.Rate = control.DefaultRate()
	sp.Bits = control.DefaultBits()
	sp.ConnTime = utils.ZeroTime
}

func isSupport(res *SpeakerResponse) bool {
	if !res.BitsMask.Combine(config.SupportAudioBits) {
		return false
	}
	if !res.RateMask.Combine(config.SupportAudioRates) {
		return false
	}

	return true
}

func updateSpeaker(sp *speaker.Speaker, support bool, res *SpeakerResponse, isFirstConn bool) error {
	sp.Supported = support

	sp.Timeout = config.OfflineValue()

	if isFirstConn {
		initSpeaker(sp, res)
	}

	sp.CheckOnline()
	pusher.Connect(sp)

	if isFirstConn {
		if !support {
			return &UnsupportError{sp}
		}

		ResponseServerInfo(sp)
	}

	control.ControlSample(sp)
	return nil
}

func CheckSpeaker(res *SpeakerResponse) (err error) {
	support := isSupport(res)

	sp := speaker.FindSpeakerByIP(res.Addr.String())

	if sp != nil {
		isFirstConn := !res.Connected

		isOnline := sp.IsOnline()

		if isFirstConn {
			// 设备重新上线，先断开先前的连接
			pusher.Disconnect(sp)
		}

		err := updateSpeaker(sp, support, res, isFirstConn)

		if !isOnline {
			bus.Trigger("speaker online", sp)
		}

		return err
	}

	sp, err = speaker.NewSpeaker(res.Addr.String(), speaker.DefaultLineID, control.DefaultChannel())
	if err != nil {
		log.Error("add speaker error", zap.Int("id", int(res.ID)))
		return err
	}

	err = updateSpeaker(sp, support, res, true)
	log.Info("found a new speaker " + sp.String())

	bus.Trigger("speaker detected", sp)

	if err != nil {
		return err
	}

	return nil
}
