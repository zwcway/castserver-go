//go:build debug
// +build debug

package api

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/decoder/localspeaker"
	"github.com/zwcway/castserver-go/decoder/pipeline"
	"github.com/zwcway/castserver-go/detector"
	"github.com/zwcway/castserver-go/pusher"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type jsonReqMessage struct {
	body []byte
}

func (j *jsonReqMessage) RequestId() string { return "" }
func (j *jsonReqMessage) Command() string   { return "" }
func (j *jsonReqMessage) Unmarshal(v any) error {
	return json.Unmarshal(j.body, v)
}

type jsonResp struct {
	Ret any
	Err string
}

func ApiDispatchDevel(ctx *fasthttp.RequestCtx) bool {
	path := string(ctx.Path())
	if len(path) < 5 || path[:5] != "/api/" {
		return false
	}
	r, ok := apiRouterList[path[5:]]
	if !ok {
		return false
	}
	req := jsonReqMessage{
		body: ctx.Request.Body(),
	}
	ret, err := r.cb(nil, &req, log)
	resp := jsonResp{
		Ret: ret,
	}
	if err != nil {
		resp.Err = err.Error()
	}
	b, err := json.Marshal(resp)
	if err != nil {
		ctx.WriteString(err.Error())
	} else {
		ctx.Write(b)
	}

	return true
}

func initDebug() {
	apiRouterList["spReconnect"] = apiRouter{apiReconnect}
	apiRouterList["sendServerInfo"] = apiRouter{apiSendServerInfo}
	apiRouterList["eventDebug"] = apiRouter{apiEventDebug}
	apiRouterList["localSpeaker"] = apiRouter{apiLocalSpeaker}
	apiRouterList["playFile"] = apiRouter{apiPlayFile}
	apiRouterList["pause"] = apiRouter{apiPlayPause}
}

func apiReconnect(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	var sp uint32
	err := req.Unmarshal(&sp)
	if err != nil {
		return nil, &Error{1, err}
	}
	s := speaker.FindSpeakerByID(speaker.ID(sp))
	if s == nil {
		return nil, &Error{4, fmt.Errorf("speaker[%d] not exists", sp)}
	}
	pusher.Disconnect(s)
	pusher.Connect(s)
	return nil, nil
}

func apiSendServerInfo(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	var spId uint32
	err := req.Unmarshal(&spId)
	if err != nil {
		return nil, err
	}

	sp := speaker.FindSpeakerByID(speaker.ID(spId))
	if sp == nil {
		return nil, nil
	}
	detector.SendServerInfo(sp)

	return nil, nil
}

func apiEventDebug(c *websockets.WSConnection, req Requester, log *zap.Logger) (ret any, err error) {
	var evt uint8

	err = req.Unmarshal(&evt)
	if err != nil {
		return
	}
	if websockets.FindEvent(websockets.Command_SPEAKER, evt) {
		sps := speaker.AllSpeakers()
		if len(sps) == 0 {
			ret = false
			return
		}
		websockets.BroadcastSpeakerEvent(sps[0], evt)
		ret = true
		return
	} else if websockets.FindEvent(websockets.Command_LINE, evt) {
		ls := speaker.LineList()
		if len(ls) == 0 {
			ret = false
			return
		}
		websockets.BroadcastLineEvent(ls[0], evt)
		ret = true
		return
	}

	switch evt {
	case websockets.Event_SP_LevelMeter:
	case websockets.Event_Line_LevelMeter:
	}

	ret = false
	return
}

func apiLocalSpeaker(c *websockets.WSConnection, req Requester, log *zap.Logger) (ret any, err error) {
	var power bool
	err = req.Unmarshal(&power)
	if err != nil {
		return
	}
	if power {
		err = localspeaker.Init()
	} else {
		localspeaker.Close()
	}

	if err != nil {
		return
	}

	return true, nil
}

func apiPlayFile(c *websockets.WSConnection, req Requester, log *zap.Logger) (ret any, err error) {
	var file struct {
		Line int
		File string
	}

	err = req.Unmarshal(&file)
	if err != nil {
		return
	}
	line := speaker.FindLineByID(speaker.LineID(file.Line))
	if line == nil {
		err = errors.New("no line")
		return
	}
	audio := pipeline.FileStreamer(line.UUID)
	err = audio.OpenFile(file.File)
	if err != nil {
		return
	}

	audio.Pause(false)
	ret = true
	return
}

func apiPlayPause(c *websockets.WSConnection, req Requester, log *zap.Logger) (ret any, err error) {
	var p struct {
		Line  int
		Pause bool
	}
	err = req.Unmarshal(&p)
	if err != nil {
		return
	}
	line := speaker.FindLineByID(speaker.LineID(p.Line))
	if line == nil {
		err = errors.New("no line")
		return
	}
	audio := pipeline.FileStreamer(line.UUID)
	audio.Pause(p.Pause)
	return true, nil
}
