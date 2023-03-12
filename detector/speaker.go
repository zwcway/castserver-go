package detector

import (
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/control"
	"github.com/zwcway/castserver-go/pusher"
	"github.com/zwcway/castserver-go/utils"
	"github.com/zwcway/castserver-go/web/websockets"

	"go.uber.org/zap"
)

func initSpeaker(sp *speaker.Speaker, res *SpeakerResponse) {
	sp.Id = res.ID
	sp.Name = res.MAC.String()
	sp.RateMask = res.RateMask
	sp.BitsMask = res.BitsMask
	sp.Dport = res.DataPort
	sp.MAC = res.MAC
	sp.IP = res.Addr
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
	go pusher.Connect(sp)

	if isFirstConn {
		if !support {
			return &UnsupportError{sp}
		}

		SendServerInfo(sp)
	}

	control.ControlSample(sp)
	return nil
}

func CheckSpeaker(res *SpeakerResponse) error {
	support := isSupport(res)

	sp := speaker.FindSpeakerByID(res.ID)

	if sp != nil {
		isFirstConn := !res.Connected

		isOnline := sp.IsOnline()

		if isFirstConn {
			// 设备重新上线，先断开先前的连接
			pusher.Disconnect(sp)
		}

		err := updateSpeaker(sp, support, res, isFirstConn)

		if !isOnline {
			// 触发设备上线事件，通知管理后台
			websockets.BroadcastSpeakerEvent(sp, websockets.Event_SP_Online)
		}

		return err
	}

	sp, err := speaker.NewSpeaker(res.ID, speaker.DefaultLineID, control.DefaultChannel())
	if err != nil {
		log.Error("add speaker error", zap.Int("id", int(res.ID)))
		return err
	}

	err = updateSpeaker(sp, support, res, true)
	log.Info("found a new speaker " + sp.String())

	// 触发设备发现事件，通知管理后台
	websockets.BroadcastSpeakerEvent(sp, websockets.Event_SP_Detected)

	if err != nil {
		return err
	}

	return nil
}
