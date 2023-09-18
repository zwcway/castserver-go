package api

import (
	"fmt"
	"unicode/utf8"

	log1 "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/web/websockets"
)

type requestLineEdit struct {
	ID              uint8   `jp:"id"`
	Name            *string `jp:"name,omitempty"`
	SpectrumLogAxis *bool   `jp:"sl,omitempty"`
}

func apiLineEdit(c *websockets.WSConnection, req Requester, log log1.Logger) (ret any, err error) {
	var p requestLineEdit
	err = req.Unmarshal(&p)
	if err != nil {
		return
	}
	nl := speaker.FindLineByID(speaker.LineID(p.ID))
	if nl == nil {
		err = fmt.Errorf("line %d not exists", p.ID)
		return
	}

	if p.Name != nil {
		if len(*p.Name) == 0 || utf8.RuneCountInString(*p.Name) > 10 {
			err = fmt.Errorf("name invalid")
			return
		}
		nl.SetName(*p.Name)
	}

	if p.SpectrumLogAxis != nil {
		nl.Input.SpectrumEle.SetLogAxis(*p.SpectrumLogAxis)
	}

	ret = true
	return
}
