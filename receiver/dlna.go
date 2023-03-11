package receiver

import (
	"github.com/zwcway/castserver-go/common/speaker"
)

func AddDLNA(line *speaker.Line) {
	if dlnaInstance == nil {
		return
	}
	line.UUID = dlnaInstance.AddNewInstance(line.Name, line.UUID)
}

func DelDLNA(line *speaker.Line) {
	if dlnaInstance == nil {
		return
	}
	dlnaInstance.DelInstance(line.UUID)
}

func EditDLNA(line *speaker.Line) {
	if dlnaInstance == nil {
		return
	}
	dlnaInstance.ChangeName(line.UUID, line.Name)
}

func SetDLNAName(uuid string, name string) {
	if dlnaInstance == nil {
		return
	}
	dlnaInstance.ChangeName(uuid, name)
}
