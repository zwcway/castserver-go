package websockets

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/go-jsonpack"
)

type notifySpeakerMoved struct {
	Speaker int `jp:"sp"`
	Type    int `jp:"type"`
	From    int `jp:"from"`
	To      int `jp:"to"`
}

func BroadcastSpeakerChannelMovedEvent(sp *speaker.Speaker, from audio.Channel, to audio.Channel) error {
	resp := notifySpeakerMoved{
		Speaker: int(sp.ID),
		Type:    1,
		From:    int(from),
		To:      int(to),
	}
	msg, err := jsonpack.Marshal(resp)
	if err != nil {
		return err
	}

	return Broadcast(Event_SP_Moved, 0, resp.Speaker, msg)
}

func BroadcastSpeakerLineMovedEvent(sp *speaker.Speaker, from speaker.LineID, to speaker.LineID) error {
	resp := notifySpeakerMoved{
		Speaker: int(sp.ID),
		Type:    2,
		From:    int(from),
		To:      int(to),
	}
	msg, err := jsonpack.Marshal(resp)
	if err != nil {
		return err
	}

	return Broadcast(Event_SP_Moved, 0, resp.Speaker, msg)
}

func BroadcastSpeakerEvent(sp *speaker.Speaker, evt uint8) error {
	msg, err := jsonpack.Marshal(NewResponseSpeakerInfo(sp))
	if err != nil {
		return err
	}
	if sp.Line != nil {
		Broadcast(Event_Line_Speaker, evt, int(sp.Line.ID), msg)
	}
	return Broadcast(evt, 0, int(sp.ID), msg)
}

func BroadcastLineEvent(line *speaker.Line, evt uint8) error {
	msg, err := jsonpack.Marshal(NewResponseLineInfo(line))
	if err != nil {
		return err
	}

	return Broadcast(evt, 0, int(line.ID), msg)
}

func BroadcastLineInputEvent(line *speaker.Line) error {
	msg, err := jsonpack.Marshal(NewResponseLineSource(line))
	if err != nil {
		return err
	}

	return Broadcast(Event_Line_Input, 0, int(line.ID), msg)
}

func Broadcast(evt uint8, sub uint8, arg int, msg []byte) error {
	// 格式： event+cmd+evt+data
	id := make([]byte, 8+len(msg))
	id[0] = 'e'
	id[1] = 'v'
	id[2] = 'e'
	id[3] = 'n'
	id[4] = 't'
	id[5] = byte(evt)
	id[6] = byte(sub)
	id[7] = byte(arg)

	for i, v := range msg {
		id[8+i] = v
	}

	// log.Debug("broadcast event",
	// 	lg.Uint8("cmd", cmd),
	// 	lg.Uint8("evt", evt),
	// 	lg.Int("arg", arg),
	// 	lg.Int("length", len(msg)),
	// )

	for c, b := range WSHub.broadcast {
		for _, e := range b {
			if e.evt == evt {
				c.Write(id)
			}
		}
	}
	return nil
}

func Subscribe(c *WSConnection, evt []uint8, sub uint8, arg int) {
	ses, ok := WSHub.broadcast[c]
	if !ok { // 设备已断开
		return
	}
	if len(evt) == 0 {
		return
	}
	addEvts := []uint8{}
	for _, ee := range evt {
		// 检查已经已经订阅过
		if findBEvent(ses, ee, sub, arg) {
			continue
		}

		// 校验事件是否合法
		if ee >= Event_MAX || ee <= Event_MIN {
			continue
		}
		// 去重
		if findEvent(addEvts, (ee)) {
			continue
		}
		addEvts = append(addEvts, (ee))
	}

	if len(addEvts) == 0 { // 所有事件都已经订阅过了
		return
	}

	// 事件为空，表示接收该cmd下的所有事件
	hasSpectrumEvent := false
	for _, e := range addEvts {
		WSHub.broadcast[c] = append(WSHub.broadcast[c], broadcastEvent{e, sub, arg})

		if isSpectrumEvent(e) {
			appendSpectrum(e, arg)
			hasSpectrumEvent = true
		}
	}

	if hasSpectrumEvent {
		startSpectumRoutine()
	}
}

func Unsubscribe(c *WSConnection, evt []uint8, sub uint8, arg int) {
	ses, ok := WSHub.broadcast[c]
	if !ok {
		return
	}
	if len(evt) == 0 {
		UnsubscribeAll(c)
		return
	}

	ne := []broadcastEvent{}

	hasSpumEvent := false

	for _, se := range ses {
		if !findEvent(evt, se.evt) || se.sub != sub {
			ne = append(ne, se)
			continue
		}

		if isSpectrumEvent(se.evt) {
			removeSpectrum(se.evt, se.arg)
			hasSpumEvent = true
		}
	}

	WSHub.broadcast[c] = ne

	if hasSpumEvent {
		if !hasSpectrumEvent(WSHub.broadcast) {
			stopSpectumRoutine()
		}
	}
}

func UnsubscribeAll(c *WSConnection) {
	if hasSpectrumEvent(WSHub.broadcast) {
		stopSpectumRoutine()
	}

	ses, ok := WSHub.broadcast[c]
	if !ok {
		return
	}

	WSHub.broadcast[c] = ses[:0]
}
