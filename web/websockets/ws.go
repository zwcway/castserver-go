package websockets

import (
	"time"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/utils"
	"go.uber.org/zap"
)

var ctx utils.Context
var log *zap.Logger
var ApiDispatch func(mt int, msg []byte, conn *websocket.Conn)

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
)

type wsConnection struct {
	hub  *Hub
	Conn *websocket.Conn
}

type Hub struct {
	// 已连接的客户端列表
	Clients map[*wsConnection]bool

	// 已订阅接收广播的客户端列表
	broadcast map[*websocket.Conn][]broadcastEvent
}

func (c *wsConnection) remoteAddr() string {
	add := c.Conn.RemoteAddr()
	if add != nil {
		return add.String()
	}
	return "unknown"
}
func (c *wsConnection) readFromClient() {
	defer func() {
		log.Debug("client close", zap.String("ip", c.Conn.RemoteAddr().String()))
		// 客户端断开
		delete(c.hub.broadcast, c.Conn)
		delete(c.hub.Clients, c)
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
		log.Debug("receive", zap.String("ip", c.Conn.RemoteAddr().String()), zap.ByteString("data", message))
		if ApiDispatch != nil {
			ApiDispatch(t, message, c.Conn)
		}
	}
	log.Error("readFromClient exited")
}

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
}

var WSHub = &Hub{
	broadcast: make(map[*websocket.Conn][]broadcastEvent),
	Clients:   make(map[*wsConnection]bool),
}

func WSHandler(ctx *fasthttp.RequestCtx) {
	err := upgrader.Upgrade(ctx, func(ws *websocket.Conn) {
		wsServer := &wsConnection{hub: WSHub, Conn: ws}

		log.Debug("client connected", zap.String("ip", ws.RemoteAddr().String()))

		// 保存客户端列表
		WSHub.Clients[wsServer] = true
		WSHub.broadcast[ws] = make([]broadcastEvent, 0)

		wsServer.readFromClient()
	})

	if err != nil {
		log.Error("ws handler error", zap.Error(err))
	}
}

func Init(c utils.Context) {
	ctx = c
	log = c.Logger("websockets")
}
