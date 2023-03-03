package receiver

import (
	"github.com/zwcway/castserver-go/common/speaker"
)

func AddDLNA(line *speaker.Line) {
	line.UUID = dlnaInstance.AddNewInstance(line.Name)
}

func DelDLNA(line *speaker.Line) {
	dlnaInstance.DelInstance(line.UUID)
}

func EditDLNA(line *speaker.Line) {
	dlnaInstance.ChangeName(line.UUID, line.Name)
}
