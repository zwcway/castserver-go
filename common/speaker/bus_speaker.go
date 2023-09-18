package speaker

import "github.com/zwcway/castserver-go/common/bus"

var (
	BusSpeakerCreated  = speakerCreated{}
	BusSpeakerEdited   = speakerEdited{}
	BusSpeakerDeleted  = speakerDeleted{}
	BusSpeakerDetected = speakerDetected{}
	BusSpeakerOnline   = speakerOnline{}
	BusSpeakerOffline  = speakerOffline{}
	BusSpeakerReonline = speakerReonline{}
)

type speakerCreated struct{}

func (speakerCreated) Dispatch(sp *Speaker) error {
	return bus.Dispatch("speaker created", sp)
}
func (speakerCreated) Register(c func(sp *Speaker) error) *bus.HandlerData {
	return bus.Register("speaker created", func(o any, a ...any) error {
		return c(a[0].(*Speaker))
	})
}

type speakerEdited struct{}

func (speakerEdited) Dispatch(sp *Speaker, a ...any) error {
	return bus.DispatchObj(sp, "speaker edited", a...)
}
func (speakerEdited) Register(c func(sp *Speaker, a ...any) error) *bus.HandlerData {
	return bus.Register("speaker edited", func(o any, a ...any) error {
		return c(o.(*Speaker), a...)
	})
}

type speakerDeleted struct{}

func (speakerDeleted) Dispatch(sp *Speaker) error {
	return bus.DispatchObj(sp, "speaker deleted")
}
func (speakerDeleted) Register(c func(sp *Speaker) error) *bus.HandlerData {
	return bus.Register("speaker deleted", func(o any, a ...any) error {
		return c(o.(*Speaker))
	})
}

type speakerDetected struct{}

func (speakerDetected) Dispatch(sp *Speaker) error {
	return bus.DispatchObj(sp, "speaker detected")
}
func (speakerDetected) Register(c func(sp *Speaker) error) *bus.HandlerData {
	return bus.Register("speaker detected", func(o any, a ...any) error {
		return c(a[0].(*Speaker))
	})
}

type speakerOffline struct{}

func (speakerOffline) Dispatch(sp *Speaker) error {
	return bus.DispatchObj(sp, "speaker offline")
}
func (speakerOffline) Register(c func(sp *Speaker) error) *bus.HandlerData {
	return bus.Register("speaker offline", func(o any, a ...any) error {
		return c(a[0].(*Speaker))
	})
}

type speakerOnline struct{}

func (speakerOnline) Dispatch(sp *Speaker) error {
	return bus.DispatchObj(sp, "speaker online")
}
func (speakerOnline) Register(c func(sp *Speaker) error) *bus.HandlerData {
	return bus.Register("speaker online", func(o any, a ...any) error {
		return c(a[0].(*Speaker))
	})
}

type speakerReonline struct{}

func (speakerReonline) Dispatch(sp *Speaker) error {
	return bus.DispatchObj(sp, "speaker reonline")
}
func (speakerReonline) Register(c func(sp *Speaker) error) *bus.HandlerData {
	return bus.Register("speaker reonline", func(o any, a ...any) error {
		return c(a[0].(*Speaker))
	})
}
