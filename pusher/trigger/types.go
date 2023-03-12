package trigger

import "github.com/zwcway/castserver-go/common/speaker"

type Trigger interface {
	AddLine(*speaker.Line)
	RemoveLine(*speaker.Line)
}
