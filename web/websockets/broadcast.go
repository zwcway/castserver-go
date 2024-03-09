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

// BroadcastSpeakerEvent 向所有客户端发送扬声器事件
func BroadcastSpeakerEvent(sp *speaker.Speaker, evt Event) error {
	msg, err := jsonpack.Marshal(NewResponseSpeakerInfo(sp))
	if err != nil {
		return err
	}
	if sp.Line != nil {
		Broadcast(Event_Line_Speaker, evt, int(sp.Line.ID), msg)
	}
	return Broadcast(evt, 0, int(sp.ID), msg)
}

// BroadcastLineEvent 广播线路事件
func BroadcastLineEvent(line *speaker.Line, evt Event) error {
	msg, err := jsonpack.Marshal(NewResponseLineInfo(line))
	if err != nil {
		return err
	}

	return Broadcast(evt, 0, int(line.ID), msg)
}

// BroadcastLineInputEvent 广播音频接入事件
func BroadcastLineInputEvent(line *speaker.Line) error {
	msg, err := jsonpack.Marshal(NewResponseLineSource(line))
	if err != nil {
		return err
	}

	return Broadcast(Event_Line_Input, 0, int(line.ID), msg)
}

// Broadcast 开始广播事件
func Broadcast(evt Event, sub Event, arg int, msg []byte) error {
	// 格式： event+cmd+evt+data
	eventMsg := make([]byte, 8+len(msg))
	eventMsg[0] = 'e'
	eventMsg[1] = 'v'
	eventMsg[2] = 'e'
	eventMsg[3] = 'n'
	eventMsg[4] = 't'
	eventMsg[5] = byte(evt)
	eventMsg[6] = byte(sub)
	eventMsg[7] = byte(arg)

	copy(eventMsg[8:], msg)

	// log.Debug("broadcast event",
	// 	lg.Uint8("cmd", cmd),
	// 	lg.Uint8("evt", evt),
	// 	lg.Int("arg", arg),
	// 	lg.Int("length", len(msg)),
	// )

	for c, evts := range WSHub.broadcast {
		for _, e := range evts {
			if e.evt == evt {
				c.Write(eventMsg)
			}
		}
	}
	return nil
}

func Subscribe(c *WSConnection, evt []Event, sub Event, arg int) {
	ses, ok := WSHub.broadcast[c]
	if !ok { // 设备已断开
		return
	}
	if len(evt) == 0 {
		return
	}
	addEvts := []Event{}
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

func Unsubscribe(c *WSConnection, evt []Event, sub Event, arg int) {
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
