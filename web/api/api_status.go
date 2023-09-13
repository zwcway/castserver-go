package api

import (
	"github.com/zwcway/castserver-go/common/config"
	log1 "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/web/websockets"
)

type requestStatus struct {
	Section string `jp:"sct"`
}

func apiStatus(c *websockets.WSConnection, req Requester, log log1.Logger) (ret any, err error) {
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

func apiStatusConfig(log log1.Logger) (ret any, err error) {
	resp := make([]responseStatusConfig, 0)

	for i := 0; i < len(config.ConfigStruct); i++ {
		for k := 0; k < len(config.ConfigStruct[i].Keys); k++ {
			val, t := config.ConfigString(&config.ConfigStruct[i], &config.ConfigStruct[i].Keys[k])

			resp = append(resp, responseStatusConfig{
				Name:  config.ConfigStruct[i].Name + "." + config.ConfigStruct[i].Keys[k].Key,
				Type:  t,
				Value: val,
				Desc:  config.ConfigStruct[i].Keys[k].Desc,
			})
		}
	}

	return resp, nil
}
