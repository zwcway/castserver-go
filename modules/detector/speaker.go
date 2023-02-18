package detector

import (
	"time"

	"github.com/zwcway/castserver-go/common/speaker"
	config "github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/modules/pusher"
	"github.com/zwcway/castserver-go/modules/web/websockets"

	"go.uber.org/zap"
)

func initSpeaker(sp *speaker.Speaker, res *Response) {
	sp.ID = res.ID
	sp.RateMask = res.RateMask
	sp.BitsMask = res.BitsMask
	sp.Dport = res.DataPort
	sp.Mport = res.CastPort
	sp.MAC = res.MAC
	sp.IP = res.Addr
}

func isSupport(res *Response) bool {
	if !res.BitsMask.IssetSlice(config.SupportAudioBits) {
		return false
	}
	if !res.RateMask.IssetSlice(config.SupportAudioRates) {
		return false
	}

	return true
}

func updateSpeaker(sp *speaker.Speaker, support bool, res *Response, isFirstConn bool) error {
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

		sp.ConnTime = uint64(time.Now().UnixMilli())

		sendServerInfo(sp)
	}

	return nil
}

func checkSpeaker(res *Response) error {
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
			websockets.BroadcastSpeakerEvent(sp, websockets.Event_SP_ONLINE)
		}

		return err
	}

	sp, err := speaker.AddSpeaker(res.ID, speaker.DefaultLine, speaker.DefaultChannel)
	if err != nil {
		log.Error("add speaker error", zap.Int("id", int(res.ID)))
		return err
	}

	err = updateSpeaker(sp, support, res, true)
	log.Info("found a new speaker " + sp.String())

	if err != nil {
		return err
	}

	// 触发设备发现事件，通知管理后台
	websockets.BroadcastSpeakerEvent(sp, websockets.Event_SP_DETECTED)

	return nil
}
