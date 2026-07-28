package element

import (
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/utils"
)

var ml lg.Logger

func Init(ctx utils.Context) error {

	ml = ctx.Logger("mixer")
	return nil
}
