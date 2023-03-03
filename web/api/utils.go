package api

import (
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/common/jsonpack"
	"go.uber.org/zap"
)

var log *zap.Logger

type apiRouter struct {
	cb func(c *websocket.Conn, req Requester, log *zap.Logger) (any, error)
}

func writeText(c *websocket.Conn, text string) {
	err := c.WriteMessage(websocket.TextMessage, []byte(text))

	if err != nil {
		log.Error("write message error", zap.Error(err))
	}
}

func writePack(c *websocket.Conn, pack any, req Requester, log *zap.Logger) {
	data, err := jsonpack.Marshal(pack)
	if err != nil {
		log.Error("marshal failed", zap.Error(err))
		return
	}
	msg := []byte(req.RequestId())
	msg = append(msg, 0)
	msg = append(msg, data...)

	err = c.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		log.Error("write message error", zap.Error(err))
	}
}
func writeError(c *websocket.Conn, err *Error, req Requester, log *zap.Logger) {
	msg := []byte(req.RequestId())
	msg = append(msg, byte(err.Code))

	log.Error("api error", zap.Int("code", int(err.Code)), zap.Error(err.Err))

	e := c.WriteMessage(websocket.BinaryMessage, msg)

	if e != nil {
		log.Error("write message error", zap.Error(err))
	}
}

type ReqMessage struct {
	id   string
	cmd  string
	body []byte
}

func (r *ReqMessage) RequestId() string {
	return r.id
}

func (r *ReqMessage) Command() string {
	return r.cmd
}

func (r *ReqMessage) Unmarshal(v any) error {
	return jsonpack.Unmarshal(r.body, v)
}

type Requester interface {
	RequestId() string
	Command() string
	Unmarshal(v any) error
}

type Error struct {
	Code uint8
	Err  error
}

func (e *Error) Error() string {
	return fmt.Sprintf("api error(%d) %s", e.Code, e.Err.Error())
}