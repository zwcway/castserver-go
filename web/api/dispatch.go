package api

import (
	"github.com/fasthttp/websocket"
	"github.com/zwcway/castserver-go/utils"
	"go.uber.org/zap"
)

var apiRouterList = map[string]apiRouter{
	"subscribe":      {apiSubscribe},
	"speakerList":    {apiSpeakerList},
	"speakerInfo":    {apiSpeakerInfo},
	"speakerVolume":  {apiSpeakerVolume},
	"setSpeaker":     {apiSpeakerEdit},
	"lineList":       {apiLineList},
	"lineInfo":       {apiLineInfo},
	"deleteLine":     {apiLineDelete},
	"createLine":     {apiLineCreate},
	"lineVolume":     {apiLineVolume},
	"linePipeLine":   {apiLinePipeLineInfo},
	"setLineEQ":      {apiLineSetEqualizer},
	"clearLineEQ":    {apiLineClearEqualizer},
	"sendServerInfo": {apiSendServerInfo},
	"spReconnect":    {apiReconnect},
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
				jp.id = string(msg[:i])
			} else {
				jp.cmd = string(msg[idx:i])
				idx = i + 1
				break
			}
		}
		if i > 30 {
			return
		}
	}
	if jp.id == "" && string(msg) == "ping" {
		apiPing(conn, &jp, log)
		return
	}
	if mt != websocket.BinaryMessage {
		return
	}
	if len(jp.cmd) <= 0 || len(jp.cmd) > 24 {
		log.Error("command invalid", zap.String("cmd", jp.cmd))
		return
	}
	jp.body = msg[idx:]

	if r, ok := apiRouterList[jp.cmd]; ok {

		ret, err := r.cb(conn, &jp, log)

		if err != nil {
			if err, ok := err.(*Error); ok {
				writeError(conn, err, &jp, log)
			} else {
				writeError(conn, &Error{1, err}, &jp, log)
			}
		} else if ret != nil {
			writePack(conn, ret, &jp, log)
		}
	}
	log.Debug("command complete", zap.String("cmd", jp.cmd))
}

func Init(ctx utils.Context) {
	log = ctx.Logger("api")
}