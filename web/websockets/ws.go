package websockets

import (
	"errors"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/utils"
	"go.uber.org/zap"
)

var ctx utils.Context
var log *zap.Logger
var ApiDispatch func(mt int, msg []byte, conn *WSConnection)

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
)

type WSConnection struct {
	hub  *Hub
	Conn *websocket.Conn
}

type writeQueueData struct {
	conn **websocket.Conn
	t    int
	data []byte
}

type broadcastMap map[*WSConnection][]broadcastEvent

type Hub struct {
	// 已连接的客户端列表
	Clients map[*WSConnection]struct{}

	// 已订阅接收广播的客户端列表
	broadcast  broadcastMap
	writeQueue chan writeQueueData
	started    bool
}

func (c *WSConnection) readFromClient() {
	defer func() {
		// 客户端断开
		UnsubscribeAll(c)
		delete(c.hub.broadcast, c)
		delete(c.hub.Clients, c)
		log.Debug("client close", zap.String("ip", c.Conn.RemoteAddr().String()))
		c.Conn.Close()
		c.Conn = nil
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		t, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("error", zap.Error(err))
			}
			break
		}
		if len(message) == 4 && string(message) == "ping" {
			WSHub.writeQueue <- writeQueueData{&c.Conn, websocket.TextMessage, []byte("pong")}
			c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			continue
		}
		// if t != websocket.TextMessage {
		// 	log.Debug("receive", zap.String("ip", c.Conn.RemoteAddr().String()), zap.ByteString("data", message))
		// }
		if ApiDispatch != nil {
			ApiDispatch(t, message, c)
		}
	}
}

func (c *WSConnection) Write(d []byte) error {
	if len(WSHub.writeQueue) >= cap(WSHub.writeQueue) {
		log.Error("write queue full", zap.Int("size", len(WSHub.writeQueue)))
		return errors.New("write queue full")
	}

	WSHub.writeQueue <- writeQueueData{&c.Conn, websocket.BinaryMessage, d}
	return nil
}

func writeToClient() {
	var d writeQueueData
	WSHub.started = true
	for {
		select {
		case <-ctx.Done():
			return
		case d = <-WSHub.writeQueue:
		}
		if *d.conn == nil {
			continue
		}
		(*d.conn).WriteMessage(d.t, d.data)
	}
}

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
}

var WSHub = &Hub{
	broadcast:  make(broadcastMap),
	Clients:    make(map[*WSConnection]struct{}),
	writeQueue: make(chan writeQueueData, 16),
}

func newConnection(ws *websocket.Conn) {
	if len(WSHub.Clients) > config.WSClientMAX {
		// 超过5个，随机断开一个
		for c := range WSHub.Clients {
			c.Conn.Close()
			break
		}
	}

	wsServer := &WSConnection{hub: WSHub, Conn: ws}

	log.Debug("client connected", zap.String("ip", ws.RemoteAddr().String()))

	// 保存客户端列表
	WSHub.Clients[wsServer] = struct{}{}
	WSHub.broadcast[wsServer] = make([]broadcastEvent, 0)

	if !WSHub.started {
		go writeToClient()
	}

	wsServer.readFromClient()
}

func WSHandler(ctx *fasthttp.RequestCtx) {
	err := upgrader.Upgrade(ctx, newConnection)

	if err != nil {
		log.Error("ws handler error", zap.Error(err))
	}
}

func Init(c utils.Context) {
	ctx = c
	log = c.Logger("websockets")

	bus.Register("speaker detected", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
		// 触发设备发现事件，通知管理后台
		BroadcastSpeakerEvent(sp, Event_SP_Detected)
		return nil
	})
	bus.Register("speaker offline", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
		BroadcastSpeakerEvent(sp, Event_SP_Offline)
		return nil
	})
	bus.Register("speaker online", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
		// 触发设备上线事件，通知管理后台
		BroadcastSpeakerEvent(sp, Event_SP_Online)
		return nil
	})
	bus.Register("speaker channel moved", func(a ...any) error {
		sp := a[0].(*speaker.Speaker)
		och := a[1].(audio.Channel)

		BroadcastSpeakerChannelMovedEvent(sp, och, sp.Channel)
		return nil
	})
	bus.Register("line volume changed", func(a ...any) error {
		line := a[0].(*speaker.Line)

		BroadcastLineEvent(line, Event_Line_Edited)
		return nil
	})
	bus.Register("line created", func(a ...any) error {
		line := a[0].(*speaker.Line)
		BroadcastLineEvent(line, Event_Line_Created)
		return nil
	})
	bus.Register("line deleted", func(a ...any) error {
		line := a[0].(*speaker.Line)
		BroadcastLineEvent(line, Event_Line_Deleted)
		return nil
	})

	bus.Register("line input changed", func(a ...any) error {
		line := a[0].(*speaker.Line)
		// 通知输入格式
		BroadcastLineInputEvent(line)
		return nil
	})
}
