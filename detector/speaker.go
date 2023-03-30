package detector

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/utils"
	"github.com/zwcway/castserver-go/control"
)

func initSpeaker(sp *speaker.Speaker, res *SpeakerResponse) {
	if len(sp.Name) == 0 {
		sp.Name = res.MAC.String()
	}
	sp.Config.RateMask = res.RateMask
	sp.Config.BitsMask = res.BitsMask
	sp.Config.AbsoluteVol = res.AbsoluteVol
	sp.Config.PowerSave = res.PowerSave
	sp.Dport = res.DataPort
	sp.Mac = res.MAC.String()
	// sp.SetFormat(audio.Format{
	// 	SampleRate: control.DefaultRate(),
	// 	SampleBits: control.DefaultBits(),
	// })
	sp.ConnTime = utils.ZeroTime
	sp.Save()
}

func isSupport(res *SpeakerResponse) bool {
	if !res.BitsMask.IntersectSlice(config.SupportAudioBits) {
		return false
	}
	if !res.RateMask.IntersectSlice(config.SupportAudioRates) {
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

	if sp.Line != nil {
		sp.Line.AppendSpeaker(sp)
	}

	sp.CheckOnline()

	if isFirstConn {
		if !support {
			return &UnsupportError{sp}
		}

		ResponseServerInfo(sp)
	}

	return nil
}

func CheckSpeaker(res *SpeakerResponse) (err error) {
	support := isSupport(res)

	sp := speaker.FindSpeakerByIP(res.Addr.String())

	if sp != nil {
		isFirstConn := !res.Connected
		isOnline := sp.IsOnline()

		err := updateSpeaker(sp, support, res, isFirstConn)

		if !isOnline || isFirstConn {
			log.Info("speaker online " + sp.String())
			bus.Dispatch("speaker online", sp)
		} else {
			log.Debug("speaker reonline " + sp.String())
			bus.Dispatch("speaker reonline", sp)
		}

		return err
	}

	sp, err = speaker.NewSpeaker(res.Addr.String(), speaker.DefaultLineID, control.DefaultChannel())
	if err != nil {
		log.Error("add speaker error", lg.Int("id", int64(res.ID)))
		return err
	}

	err = updateSpeaker(sp, support, res, true)
	log.Info("found a new speaker " + sp.String())

	bus.Dispatch("speaker detected", sp)

	if err != nil {
		return err
	}

	return nil
}
