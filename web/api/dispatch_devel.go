//go:build debug
// +build debug

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/netip"

	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/common/audio"
	log1 "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/pipeline"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/control"
	"github.com/zwcway/castserver-go/decoder"
	"github.com/zwcway/castserver-go/decoder/localspeaker"
	"github.com/zwcway/castserver-go/detector"
	"github.com/zwcway/castserver-go/pusher"
	"github.com/zwcway/castserver-go/web/websockets"
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
	apiRouterList["addSpeaker"] = apiRouter{apiSpeakerCreate}
	apiRouterList["spReconnect"] = apiRouter{apiReconnect}
	apiRouterList["spSample"] = apiRouter{apiControlSpeakerFormat}
	apiRouterList["sendServerInfo"] = apiRouter{apiSendServerInfo}
	apiRouterList["eventDebug"] = apiRouter{apiEventDebug}
	apiRouterList["localSpeaker"] = apiRouter{apiLocalSpeaker}
	apiRouterList["playFile"] = apiRouter{apiPlayFile}
	apiRouterList["pause"] = apiRouter{apiPlayPause}
	apiRouterList["debugStatus"] = apiRouter{apiDebugStatus}
	apiRouterList["elementPower"] = apiRouter{apiElementPower}
}

type requestSpeakerCreate struct {
	Ver      uint8
	ID       uint32
	IP       string
	MAC      string
	DataPort uint16
	BitsMask []uint8
	RateMask []uint8
	AVol     bool
}

func apiSpeakerCreate(c *websockets.WSConnection, req Requester, log log1.Logger) (any, error) {
	p := requestSpeakerCreate{}
	if err := req.Unmarshal(&p); err != nil {
		return nil, err
	}

	mac, err := net.ParseMAC(p.MAC)
	if err != nil {
		return nil, err
	}
	rm, err := audio.NewAudioRateMask(p.RateMask)
	if err != nil {
		return nil, err
	}
	bm, err := audio.NewAudioBitsMask(p.BitsMask)
	if err != nil {
		return nil, err
	}
	res := &detector.SpeakerResponse{
		Ver:         uint8(p.Ver),
		ID:          speaker.SpeakerID(p.ID),
		Connected:   false,
		Addr:        netip.MustParseAddr(p.IP),
		MAC:         mac,
		RateMask:    rm,
		BitsMask:    bm,
		DataPort:    p.DataPort,
		AbsoluteVol: p.AVol,
		PowerSave:   true,
	}
	detector.CheckSpeaker(res)

	return true, nil
}

func apiReconnect(c *websockets.WSConnection, req Requester, log log1.Logger) (any, error) {
	var sp uint32
	err := req.Unmarshal(&sp)
	if err != nil {
		return nil, &Error{1, err}
	}
	s := speaker.FindSpeakerByID(speaker.SpeakerID(sp))
	if s == nil {
		return nil, &Error{4, fmt.Errorf("speaker[%d] not exists", sp)}
	}
	pusher.Disconnect(s)
	pusher.Connect(s)
	return true, nil
}

func apiSendServerInfo(c *websockets.WSConnection, req Requester, log log1.Logger) (any, error) {
	var spId uint32
	err := req.Unmarshal(&spId)
	if err != nil {
		return nil, err
	}

	sp := speaker.FindSpeakerByID(speaker.SpeakerID(spId))
	if sp == nil {
		return nil, nil
	}
	detector.ResponseServerInfo(sp)

	return true, nil
}

func apiControlSpeakerFormat(c *websockets.WSConnection, req Requester, log log1.Logger) (any, error) {
	var spId uint32
	err := req.Unmarshal(&spId)
	if err != nil {
		return nil, err
	}

	sp := speaker.FindSpeakerByID(speaker.SpeakerID(spId))
	if sp == nil {
		return nil, nil
	}
	control.ControlSample(sp)

	return true, nil
}

func apiEventDebug(c *websockets.WSConnection, req Requester, log log1.Logger) (ret any, err error) {
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

func apiLocalSpeaker(c *websockets.WSConnection, req Requester, log log1.Logger) (ret any, err error) {
	var power bool
	err = req.Unmarshal(&power)
	if err != nil {
		return
	}
	if power {
		err = localspeaker.Init()
		for _, line := range speaker.LineList() {
			localspeaker.AddLine(line)
		}
		localspeaker.Play()
	} else {
		localspeaker.Close()
	}

	if err != nil {
		return
	}

	return true, nil
}

func apiPlayFile(c *websockets.WSConnection, req Requester, log log1.Logger) (ret any, err error) {
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
	audio := decoder.FileStreamer(line.UUID)
	err = audio.OpenFile(file.File)
	if err != nil {
		return
	}

	audio.SetPause(false)
	ret = true
	return
}

func apiPlayPause(c *websockets.WSConnection, req Requester, log log1.Logger) (ret any, err error) {
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
	audio := decoder.FileStreamer(line.UUID)
	audio.SetPause(p.Pause)
	return true, nil
}

func apiDebugStatus(c *websockets.WSConnection, req Requester, log log1.Logger) (ret any, err error) {
	var p struct {
		Line *int `jp:"line,omitempty"`
	}
	err = req.Unmarshal(&p)
	if err != nil {
		return
	}
	type r struct {
		Name  string `jp:"name"`
		Power int    `jp:"on"`
		Cost  int    `jp:"cost"`
	}

	resp := struct {
		LocalSpeaker bool   `jp:"local"`
		LocalPlaying bool   `jp:"lplay"`
		FilePlaying  bool   `jp:"fplay"`
		FileName     string `jp:"furl"`
		SpectrumLog  bool   `jp:"sl"`
		Elements     []r    `jp:"eles"`
	}{
		LocalSpeaker: localspeaker.IsOpened(),
		LocalPlaying: localspeaker.IsPlaying(),
	}

	var pl *pipeline.PipeLine
	if p.Line != nil {
		line := speaker.FindLineByID(speaker.LineID(*p.Line))
		if line == nil {
			err = errors.New("no line")
			return
		}
		fs := line.Input.FileStreamer()
		if fs != nil {
			resp.FilePlaying = !fs.IsPaused()
			resp.FileName = fs.CurrentFile()
		}
		resp.SpectrumLog = line.SpectrumEle.LogAxis()
		if line.Input.PipeLine != nil {
			pl, _ = line.Input.PipeLine.(*pipeline.PipeLine)
		}
	}

	if pl != nil {
		for _, s := range pl.Streamers() {
			res := r{
				Name: s.Name(),
				Cost: int(s.Cost().Microseconds()),
			}
			ele := s.Element()
			if ele, ok := ele.(stream.SwitchElement); ok {
				if ele.IsOn() {
					res.Power = 1
				} else {
					res.Power = 0
				}
			} else {
				res.Power = -1
			}

			resp.Elements = append(resp.Elements, res)
		}
	}

	return resp, nil
}

func apiElementPower(c *websockets.WSConnection, req Requester, log log1.Logger) (ret any, err error) {
	var p struct {
		Line    *int   `jp:"line,omitempty"`
		Speaker *int   `jp:"sp,omitempty"`
		Name    string `jp:"n"`
		Power   *bool  `jp:"on"`
	}
	err = req.Unmarshal(&p)
	if err != nil {
		return
	}
	var pl stream.PipeLiner

	if p.Line != nil {
		line := speaker.FindLineByID(speaker.LineID(*p.Line))
		if line == nil {
			err = errors.New("no line")
			return
		}
		pl = line.Input.PipeLine
	}

	if pl == nil {
		err = errors.New("no pipeline")
		return
	}
	pp, ok := pl.(*pipeline.PipeLine)
	if !ok {
		err = errors.New("unknown pipeline")
		return
	}
	var ele stream.Element
	for _, s := range pp.Streamers() {
		if s.Name() == p.Name {
			ele = s.Element()
			break
		}
	}
	if ele == nil {
		err = errors.New("no element")
		return
	}
	if p.Power != nil {
		if ele, ok := ele.(stream.SwitchElement); ok {
			if *p.Power {
				ele.On()
			} else {
				ele.Off()
			}
		}
	}

	return true, nil
}
