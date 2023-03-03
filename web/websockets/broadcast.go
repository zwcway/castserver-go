package websockets

import (
	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/jsonpack"
	"github.com/zwcway/castserver-go/common/speaker"
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

	return Broadcast(Command_SPEAKER, Event_SP_Moved, resp.Speaker, msg)
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

	return Broadcast(Command_SPEAKER, Event_SP_Moved, resp.Speaker, msg)
}

func BroadcastSpeakerEvent(sp *speaker.Speaker, evt uint8) error {
	msg, err := jsonpack.Marshal(sp)
	if err != nil {
		return err
	}

	return Broadcast(Command_SPEAKER, evt, int(sp.ID), msg)
}

func BroadcastLineEvent(line *speaker.Line, evt uint8) error {
	msg, err := jsonpack.Marshal(line)
	if err != nil {
		return err
	}

	return Broadcast(Command_LINE, evt, int(line.ID), msg)
}

func Broadcast(cmd, evt uint8, arg int, msg []byte) error {
	// 格式： event+cmd+evt+data
	id := make([]byte, 8+len(msg))
	id[0] = 'e'
	id[1] = 'v'
	id[2] = 'e'
	id[3] = 'n'
	id[4] = 't'
	id[5] = byte(cmd)
	id[6] = byte(evt)
	id[7] = byte(arg)

	for i, v := range msg {
		id[7+i] = v
	}

	for c, b := range WSHub.broadcast {
		for _, e := range b {
			if e.evt != evt {
				continue
			}

			c.WriteMessage(websocket.BinaryMessage, msg)
		}
	}
	return nil
}

func Subscribe(c *websocket.Conn, cmd uint8, evt []uint8, arg int, hs map[uint8]EventHandler) {
	ses, ok := WSHub.broadcast[c]
	if !ok { // 设备已断开
		return
	}
	addEvts := []uint8{}
	for _, ee := range evt {
		// 检查已经已经订阅过
		if findBEvent(ses, (ee)) {
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
	if len(evt) == 0 {
		es, ok := CommandEventMap[cmd]
		if !ok {
			return
		}
		for _, e := range es {
			if findBEvent(ses, e) {
				continue
			}
			addEvts = append(addEvts, e)
		}
	}

	if len(addEvts) == 0 { // 所有事件都已经订阅过了
		return
	}

	// 事件为空，表示接收该cmd下的所有事件
	for _, e := range addEvts {
		WSHub.broadcast[c] = append(WSHub.broadcast[c], broadcastEvent{e, arg})
		if h, ok := hs[e]; ok {
			h.On(e, arg, ctx, log)
		}
	}
}

func Unsubscribe(c *websocket.Conn, cmd uint8, evt []uint8, arg int, hs map[uint8]EventHandler) {
	ses, ok := WSHub.broadcast[c]
	if !ok {
		return
	}

	ne := []broadcastEvent{}

	for _, ee := range evt {
		if findBEvent(ses, (ee)) {
			if h, ok := hs[ee]; ok {
				h.Off(ee, arg)
			}
			continue
		}
		if ee >= Event_MAX || ee <= Event_MIN {
			continue
		}
		ne = append(ne, broadcastEvent{ee, arg})
	}

	WSHub.broadcast[c] = ne
}
