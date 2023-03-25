package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestLinePlayerSeek struct {
	ID  uint8 `jp:"id"`
	Pos int   `jp:"pos"`
}

func apiLinePlayerSeek(c *websockets.WSConnection, req Requester, log *zap.Logger) (any, error) {
	var p requestLinePlayerSeek
	err := req.Unmarshal(&p)
	if err != nil {
		return nil, err
	}

	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		return nil, fmt.Errorf("line not exists")
	}

	fs := nl.Input.FileStreamer()
	if fs == nil {
		return nil, errors.New("no audio")
	}

	pos := time.Duration(p.Pos) * time.Second
	err = fs.Seek(pos)
	if err != nil {
		return nil, err
	}

	bus.Trigger("line audiofile seek", nl, fs, pos)

	return websockets.NewResponseLineSource(nl), nil
}
