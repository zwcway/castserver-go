package pipeline

import (
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/utils"
)

var log lg.Logger

func Init(ctx utils.Context) error {
	log = ctx.Logger("pipeline")

	return nil
}
