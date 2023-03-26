package common

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/utils"
)

var (
	Module = commonModule{}
)

type commonModule struct{}

func (commonModule) Init(ctx utils.Context) error {
	bus.Init(ctx)
	return speaker.Init()
}

func (commonModule) DeInit() {

}

func LoadData() error {
	return speaker.LoadData()
}
