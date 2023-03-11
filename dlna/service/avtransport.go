package service

import (
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/decoder/localspeaker"
	"github.com/zwcway/castserver-go/decoder/pipeline"
	"github.com/zwcway/castserver-go/utils"
	"github.com/zwcway/fasthttp-upnp/avtransport1"
	"github.com/zwcway/fasthttp-upnp/soap"
	"go.uber.org/zap"
)

var playUri string
var metaData string

func setAVTransportURIHandler(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	in := input.(*avtransport1.ArgInSetAVTransportURI)
	// out := output.(*avtransport1.ArgOutSetAVTransportURI)

	playUri = in.CurrentURI
	metaData = in.CurrentURIMetaData

	audio := pipeline.FileStreamer(uuid)
	audio.Close()

	var err error
	if err = audio.OpenFile(playUri); err != nil {
		log.Error("create decoder failed", zap.Error(err))
		return &soap.Error{Code: 500, Desc: err.Error()}
	}
	localspeaker.Init()

	log.Info("set uri", zap.String("url", in.CurrentURI), zap.String("format", audio.AudioFormat().String()))

	return nil
}

func avtGetPositionInfo(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	out := output.(*avtransport1.ArgOutGetPositionInfo)

	audio := pipeline.FileStreamer(uuid)

	dur := utils.FormatDuration(audio.Duration())
	out.TrackDuration = dur
	out.TrackURI = audio.CurrentFile()
	out.RelTime = dur
	out.AbsTime = dur
	out.Track = 0
	out.TrackMetaData = metaData
	out.AbsCount = int32(audio.Position())
	out.RelCount = int32(audio.Position())

	return nil
}

func avtPlay(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	// in := input.(*avtransport1.ArgInPlay)

	if playUri == "" {
		return nil
	}
	audio := pipeline.FileStreamer(uuid)

	if audio.CurrentFile() == "" { // 重新播放
		var err error
		if err = audio.OpenFile(playUri); err != nil {
			log.Error("create decoder failed", zap.Error(err))
			return &soap.Error{Code: 500, Desc: err.Error()}
		}
		localspeaker.Init()
	}
	audio.Pause(false)

	return nil
}

func avtPause(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	if playUri == "" {
		return nil
	}
	pipeline.FileStreamer(uuid).Pause(true)

	return nil
}

func avtStop(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	if playUri == "" {
		return nil
	}
	pipeline.FileStreamer(uuid).Close()
	return nil
}

func avtSeek(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	in := input.(*avtransport1.ArgInSeek)

	if playUri == "" {
		return nil
	}
	switch in.Unit {
	case "ABS_TIME", "REL_TIME":
		d, err := utils.ParseDuration(in.Target)
		if err != nil {
			return &soap.Error{Code: fasthttp.StatusBadRequest, Desc: err.Error()}
		}
		err = pipeline.FileStreamer(uuid).Seek(d)
		if err != nil {
			return &soap.Error{Code: fasthttp.StatusBadRequest, Desc: err.Error()}
		}
	}

	return nil
}

func parseSpeed(sp string) float64 {
	sl := strings.Split(sp, "/")
	if len(sl) == 1 {
		i, err := strconv.Atoi(sp)
		if err != nil {
			return 1.0
		}
		return float64(i)
	} else if len(sl) == 2 {
		num, err := strconv.Atoi(sl[0])
		if err != nil {
			return 1.0
		}
		den, err := strconv.Atoi(sl[1])
		if err != nil {
			return 1.0
		}
		return float64(num) / float64(den)
	}
	return 1.0
}
