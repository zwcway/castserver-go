package speaker

import (
	"sync"

	"github.com/zwcway/castserver-go/common/bus"
)

var locker sync.Mutex

func Init() error {
	return nil
}

func LoadData() error {
	err := initLine()
	if err != nil {
		return err
	}
	return nil
}

func initLine() error {
	lineList = lineList[:0]
	err := BusGetLines.Dispatch(&lineList)
	if err != nil {
		return err
	}

	line := FindLineByID(DefaultLineID)
	if line == nil {
		maxLineID = 1
		NewLine("Default")
		return nil
	}
	speakerList = speakerList[:0]
	err = bus.Dispatch("get speakers", &speakerList)
	if err != nil {
		return err
	}
	linkLineAndSpeaker()
	return nil
}

func linkLineAndSpeaker() {
	for _, line := range lineList {
		if line.ID > maxLineID {
			maxLineID = line.ID
		}
		line.init()
		for _, sp := range speakerList {
			if sp.LineId == line.ID {
				sp.Line = line
				sp.init()

				line.AppendSpeaker(sp)

				bus.Dispatch("speaker created", sp)
			}
		}
		bus.Dispatch("line created", line)
	}
_sp_:
	for _, sp := range speakerList {
		for _, line := range lineList {
			if sp.LineId == line.ID {
				continue _sp_
			}
		}
		sp.init()
		bus.Dispatch("speaker created", sp)
	}
}
