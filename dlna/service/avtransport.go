package service

import (
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/decoder"
	"github.com/zwcway/castserver-go/decoder/localspeaker"
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

	audioFS := decoder.FileStreamer(uuid)

	var err error
	if err = audioFS.OpenFile(playUri); err != nil {
		log.Error("create decoder failed", zap.Error(err))
		return &soap.Error{Code: 500, Desc: err.Error()}
	}

	log.Info("set uri", zap.String("url", in.CurrentURI), zap.String("format", audioFS.AudioFormat().String()))

	return nil
}

func avtGetPositionInfo(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	out := output.(*avtransport1.ArgOutGetPositionInfo)

	audioFS := decoder.FileStreamer(uuid)

	dur := utils.FormatDuration(audioFS.Duration())
	out.TrackDuration = dur
	out.TrackURI = audioFS.CurrentFile()
	out.RelTime = dur
	out.AbsTime = dur
	out.Track = 0
	out.TrackMetaData = metaData
	out.AbsCount = int32(audioFS.Position())
	out.RelCount = int32(audioFS.Position())

	return nil
}

func avtPlay(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	// in := input.(*avtransport1.ArgInPlay)

	if playUri == "" {
		return nil
	}
	audioFS := decoder.FileStreamer(uuid)

	if audioFS.CurrentFile() == "" { // 重新播放
		var err error
		if err = audioFS.OpenFile(playUri); err != nil {
			log.Error("create decoder failed", zap.Error(err))
			return &soap.Error{Code: 500, Desc: err.Error()}
		}
		localspeaker.Init()
	}
	audioFS.SetPause(false)

	return nil
}

func avtPause(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	if playUri == "" {
		return nil
	}
	decoder.FileStreamer(uuid).SetPause(true)

	return nil
}

func avtStop(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	if playUri == "" {
		return nil
	}
	decoder.FileStreamer(uuid).Close()
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
		err = decoder.FileStreamer(uuid).Seek(d)
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
