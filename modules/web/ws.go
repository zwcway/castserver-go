package web

import (
	"time"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type wsConnection struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

type Hub struct {
	// 已连接的客户端列表
	clients map[*wsConnection]bool

	// 已订阅接收广播的客户端列表
	broadcast map[*wsConnection]bool
}

func (c *wsConnection) readFromClient() {
	defer func() {
		// 客户端断开
		delete(c.hub.broadcast, c)
		delete(c.hub.clients, c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		t, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("error", zap.Error(err))
			}
			break
		}
		apiDispatch(t, message, c.conn)
	}
}

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		return true
	},
}

var wsHub = &Hub{
	broadcast: make(map[*wsConnection]bool),
	clients:   make(map[*wsConnection]bool),
}

func BroadCast() {

}

func wsHandler(ctx *fasthttp.RequestCtx) {

	err := upgrader.Upgrade(ctx, func(ws *websocket.Conn) {
		wsServer := &wsConnection{hub: wsHub, conn: ws, send: make(chan []byte, 256)}

		// 保存客户端列表
		wsHub.clients[wsServer] = true

		wsServer.readFromClient()
	})

	if err != nil {
		log.Error("ws handler error", zap.Error(err))
	}
}
