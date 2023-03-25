package speaker

import "sync"

var locker sync.Mutex

func Init() error {
	err := initLine()
	if err != nil {
		return err
	}
	return initSpeaker()
}
