package common

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/utils"
)

func Init(ctx utils.Context) error {
	bus.Init(ctx)
	return speaker.Init()
}

func Deinit() error {

	return nil
}
