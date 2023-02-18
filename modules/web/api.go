package web

import (
	"fmt"
	"strings"

	"github.com/zwcway/castserver-go/common/jsonpack"
	"github.com/zwcway/castserver-go/modules/web/api"

	"github.com/fasthttp/websocket"
	"go.uber.org/zap"
)

var apiRouterList map[string]apiRouter = map[string]apiRouter{
	"ping":           {false, true, api.Ping},
	"speakerList":    {true, false, api.SpeakerList},
	"sendServerInfo": {false, true, api.SendServerInfo},
}

type apiRouter struct {
	allowBinary bool
	allowText   bool
	cb          func(c *websocket.Conn, req any)
}

type ReqMessage struct {
	RequestId string `jp:"id"`
	Command   string `jp:"cmd"`
	Params    any    `jp:"params"`
}

func (m *ReqMessage) check() bool {
	if len(m.Command) <= 0 || len(m.Command) > 24 {
		log.Error("command invalid", zap.String("cmd", m.Command))
		return false
	}
	return true
}

func apiDispatch(mt int, msg []byte, conn *websocket.Conn) {
	jp := ReqMessage{}
	if mt == websocket.BinaryMessage {
		err := jsonpack.Unmarshal(msg, &jp)
		if err != nil {
			log.Error("request invalid", zap.Binary("msg", msg))
			return
		}
	} else {
		list := strings.Split(string(msg), "\x00")
		if len(list) != 2 {
			log.Error("request invalid", zap.Binary("msg", msg))
			return
		}
		jp.Command = list[0]
		jp.Params = list[1]
	}
	if !jp.check() {
		return
	}

	if r, ok := apiRouterList[jp.Command]; ok {
		r.cb(conn, jp.Params)
	}
	fmt.Println(mt, msg)
}
