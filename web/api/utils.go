package api

import (
	"fmt"

	log1 "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/web/websockets"
	"github.com/zwcway/go-jsonpack"
)

var log log1.Logger

type apiRouter struct {
	cb func(c *websockets.WSConnection, req Requester, log log1.Logger) (any, error)
}

func writePack(c *websockets.WSConnection, pack any, req Requester, log log1.Logger) {
	data, err := jsonpack.Marshal(pack)
	if err != nil {
		log.Error("marshal failed", log1.Error(err))
		return
	}
	msg := []byte(req.RequestId())
	msg = append(msg, 0)
	msg = append(msg, data...)

	err = c.Write(msg)
	if err != nil {
		log.Error("write message error", log1.Error(err))
	}
}
func writeError(c *websockets.WSConnection, err *Error, req Requester, log log1.Logger) {
	msg := []byte(req.RequestId())
	msg = append(msg, byte(err.Code))

	log.Error(req.Command()+" api error",
		log1.Int("code", int64(err.Code)),
		log1.Error(err.Err),
		log1.String("reqid", req.RequestId()),
	)

	e := c.Write(msg)

	if e != nil {
		log.Error("write message error", log1.Error(err))
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
