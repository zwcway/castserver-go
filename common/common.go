package common

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/database"
	"github.com/zwcway/castserver-go/common/element"
	"github.com/zwcway/castserver-go/common/pipeline"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/common/utils"
)

func Init(ctx utils.Context) error {
	bus.Init(ctx)

	database.Init(ctx, config.DB)

	if err := speaker.Init(ctx); err != nil {
		return err
	}
	if err := element.Init(ctx); err != nil {
		return err
	}
	if err := pipeline.Init(ctx); err != nil {
		return err
	}

	speaker.LoadData()

	return nil
}
