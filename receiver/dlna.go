package receiver

import (
	"github.com/zwcway/castserver-go/common/config"
	"github.com/zwcway/castserver-go/common/speaker"
	"github.com/zwcway/castserver-go/receiver/dlna"
)

func initDlna() error {
	var err error

	if !config.EnableDLNA {
		return nil
	}

	dlnaInstance, err = dlna.NewDLNAServer(ctx)
	if err != nil {
		return err
	}
	go dlnaInstance.ListenAndServe()

	speaker.BusLineNameChanged.Register(EditDLNA)
	speaker.BusLineCreated.Register(AddDLNA)
	speaker.BusLineDeleted.Register(DelDLNA)

	return nil
}

func AddDLNA(line *speaker.Line) error {
	if dlnaInstance != nil {
		line.UUID = dlnaInstance.AddNewInstance(line.LineName, line.UUID)
	}
	return nil
}

func DelDLNA(line, dst *speaker.Line) error {
	if dlnaInstance != nil {
		dlnaInstance.DelInstance(line.UUID)
	}
	return nil
}

func EditDLNA(line *speaker.Line, old *string) error {
	if dlnaInstance != nil {
		dlnaInstance.ChangeName(line.UUID, line.LineName)
	}
	return nil
}

func SetDLNAName(uuid string, name string) {
	if dlnaInstance == nil {
		return
	}
	dlnaInstance.ChangeName(uuid, name)
}
