package service

import (
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/zwcway/castserver-go/decoder"
	"github.com/zwcway/fasthttp-upnp/avtransport1"
	"github.com/zwcway/fasthttp-upnp/soap"
	"go.uber.org/zap"
)

var playUri string
var metaData string

func setAVTransportURIHandler(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	in := input.(*avtransport1.ArgInSetAVTransportURI)
	// out := output.(*avtransport1.ArgOutSetAVTransportURI)

	log.Info("set uri", zap.String("url", in.CurrentURI), zap.Int("rate", audioDecoder.SampleRate()), zap.Int("channels", audioDecoder.Channels()))

	playUri = in.CurrentURI
	metaData = in.CurrentURIMetaData

	audioDecoder.Close()
	var err error
	if err = audioDecoder.Decode(playUri); err != nil {
		log.Error("create decoder failed", zap.Error(err))
		return &soap.Error{Code: 500, Desc: err.Error()}
	}
	return nil
}

func avtGetPositionInfo(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	out := output.(*avtransport1.ArgOutGetPositionInfo)

	dur := decoder.DurationFormat(audioDecoder.Duration())
	out.TrackDuration = dur
	out.TrackURI = audioDecoder.CurrentFile()
	out.RelTime = dur
	out.AbsTime = dur
	out.Track = 0
	out.TrackMetaData = metaData
	out.AbsCount = int32(audioDecoder.Position())
	out.RelCount = int32(audioDecoder.Position())

	return nil
}

func avtPlay(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	in := input.(*avtransport1.ArgInPlay)

	speed := parseSpeed(in.Speed)

	if playUri == "" {
		return nil
	}
	if audioDecoder.CurrentFile() == "" { // 重新播放
		var err error
		if err = audioDecoder.Decode(playUri); err != nil {
			log.Error("create decoder failed", zap.Error(err))
			return &soap.Error{Code: 500, Desc: err.Error()}
		}
	}

	if audioDecoder.IsPaused() {
		audioDecoder.Unpause()
	} else {
		audioDecoder.LocalPlay()
	}
	audioDecoder.SetSpeed(speed)

	return nil
}

func avtPause(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	if playUri == "" {
		return nil
	}
	audioDecoder.Pause()

	return nil
}

func avtStop(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	if playUri == "" {
		return nil
	}

	audioDecoder.Close()
	return nil
}

func avtSeek(input any, output any, ctx *fasthttp.RequestCtx, uuid string) error {
	in := input.(*avtransport1.ArgInSeek)

	if playUri == "" {
		return nil
	}
	switch in.Unit {
	case "ABS_TIME", "REL_TIME":
		d, err := decoder.ParseDuration(in.Target)
		if err != nil {
			return &soap.Error{Code: fasthttp.StatusBadRequest, Desc: err.Error()}
		}
		err = audioDecoder.Seek(d)
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
