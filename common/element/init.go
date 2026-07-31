package element

import (
	lg "github.com/zwcway/castserver-go/common/log"
	"github.com/zwcway/castserver-go/common/utils"
)

var (
	ml  lg.Logger
	ctx utils.Context
)

func Init(c utils.Context) error {
	ctx = c
	ml = c.Logger("mixer")
	return nil
}
