package api

import (
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/web/websockets"
	"go.uber.org/zap"
)

type requestStatus struct {
	Section string `jp:"sct"`
}

func apiStatus(c *websockets.WSConnection, req Requester, log *zap.Logger) (ret any, err error) {
	var p requestStatus
	err = req.Unmarshal(&p)
	if err != nil {
		return
	}
	switch p.Section {
	case "config":
		return apiStatusConfig(log)
	}

	return
}

type responseStatusConfig struct {
	Name  string `jp:"name"`
	Type  string `jp:"t"`
	Value string `jp:"val"`
	Desc  string `jp:"desc"`
}

func apiStatusConfig(log *zap.Logger) (ret any, err error) {
	resp := make([]responseStatusConfig, 0)

	for _, cs := range config.ConfigStruct {
		for _, ck := range cs.Keys {
			val, t := config.ConfigString(&cs, &ck)

			resp = append(resp, responseStatusConfig{
				Name:  cs.Name + "." + ck.Key,
				Type:  t,
				Value: val,
				Desc:  ck.Desc,
			})
		}
	}

	return resp, nil
}
