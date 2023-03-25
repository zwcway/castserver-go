package receiver

import (
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/config"
	"github.com/zwcway/castserver-go/receiver/dlna"
)

func initDlna() error {
	var err error

	if !config.EnableDLNA {
		return nil
	}

	line := speaker.DefaultLine()

	dlnaInstance, line.UUID, err = dlna.NewDLNAServer(ctx, line.Name)
	if err != nil {
		return err
	}
	go dlnaInstance.ListenAndServe()

	bus.Register("line name edited", func(a ...any) error {
		line := a[0].(*speaker.Line)

		EditDLNA(line)
		return nil
	})
	return nil
}

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
