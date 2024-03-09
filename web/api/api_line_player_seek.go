package api

import (
	"errors"
	"fmt"
	"time"

	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

type requestLinePlayerSeek struct {
	ID  uint8 `jp:"id"`
	Pos int   `jp:"pos"`
}

func apiLinePlayerSeek(c *websockets.WSConnection, req Requester, log lg.Logger) (any, error) {
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

	nl.Dispatch("line audiofile seek", fs, pos)

	return websockets.NewResponseLineSource(nl), nil
}
