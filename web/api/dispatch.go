package api

import (
	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/utils"
	"go.uber.org/zap"
)

var apiRouterList map[string]apiRouter = map[string]apiRouter{
	"subscribe":      {false, true, apiSubscribe},
	"speakerList":    {true, false, apiSpeakerList},
	"speakerInfo":    {true, false, apiSpeakerInfo},
	"speakerVolume":  {false, true, apiSpeakerVolume},
	"setChannel":     {false, true, apiSpeakerSetChannel},
	"lineList":       {false, true, apiLineList},
	"lineInfo":       {false, true, apiLineInfo},
	"deleteLine":     {false, true, apiLineDelete},
	"createLine":     {false, true, apiLineCreate},
	"lineVolume":     {false, true, apiLineVolume},
	"setLineEQ":      {false, true, apiLineSetEqualizer},
	"clearLineEQ":    {false, true, apiLineClearEqualizer},
	"sendServerInfo": {false, true, apiSendServerInfo},
	"spReconnect":    {false, true, apiReconnect},
}

func ApiDispatch(mt int, msg []byte, conn *websocket.Conn) {
	var (
		jp  = ReqMessage{}
		idx = 0
	)

	for i, b := range msg {
		if b == 0 {
			if idx == 0 {
				idx = i + 1
				jp.RequestId = string(msg[:i])
			} else {
				jp.Command = string(msg[idx:i])
				idx = i + 1
				break
			}
		}
		if i > 30 {
			return
		}
	}
	if jp.RequestId == "" && string(msg) == "ping" {
		apiPing(conn, &jp, log)
		return
	}
	if mt != websocket.BinaryMessage {
		return
	}
	if len(jp.Command) <= 0 || len(jp.Command) > 24 {
		log.Error("command invalid", zap.String("cmd", jp.Command))
		return
	}
	jp.Req = msg[idx:]

	if r, ok := apiRouterList[jp.Command]; ok {

		ret, err := r.cb(conn, &jp, log)

		if err != nil {
			writeError(conn, &Error{1, err}, &jp, log)
		} else if ret != nil {
			writePack(conn, ret, &jp, log)
		}
	}
	log.Debug("command complete", zap.String("cmd", jp.Command))
}

func Init(ctx utils.Context) {
	log = ctx.Logger("api")
}
