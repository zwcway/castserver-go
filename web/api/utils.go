package api

import (
	"fmt"

	"github.com/zwcway/castserver-go/common/jsonpack"
	"github.com/zwcway/castserver-go/common/lg"
	"github.com/zwcway/castserver-go/web/websockets"
)

var log lg.Logger

type apiRouter struct {
	cb func(c *websockets.WSConnection, req Requester, log lg.Logger) (any, error)
}

func writePack(c *websockets.WSConnection, pack any, req Requester, log lg.Logger) {
	data, err := jsonpack.Marshal(pack)
	if err != nil {
		log.Error("marshal failed", lg.Error(err))
		return
	}
	msg := []byte(req.RequestId())
	msg = append(msg, 0)
	msg = append(msg, data...)

	err = c.Write(msg)
	if err != nil {
		log.Error("write message error", lg.Error(err))
	}
}
func writeError(c *websockets.WSConnection, err *Error, req Requester, log lg.Logger) {
	msg := []byte(req.RequestId())
	msg = append(msg, byte(err.Code))

	log.Error(req.Command()+" api error",
		lg.Int("code", int64(err.Code)),
		lg.Error(err.Err),
		lg.String("reqid", req.RequestId()),
	)

	e := c.Write(msg)

	if e != nil {
		log.Error("write message error", lg.Error(err))
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
