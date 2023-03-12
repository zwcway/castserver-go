package speaker

func Init() error {
	err := initLine()
	if err != nil {
		return err
	}
	return initSpeaker()
}
