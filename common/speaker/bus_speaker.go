package speaker

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
)

var (
	BusGetSpeakers          = getSpeakers{}
	BusGetSpeaker           = getSpeaker{}
	BusSpeakerCreated       = speakerCreated{}
	BusSpeakerEdited        = speakerEdited{}
	BusSpeakerSaved         = speakerSaved{}
	BusSpeakerDeleted       = speakerDeleted{}
	BusSpeakerDetected      = speakerDetected{}
	BusSpeakerOnline        = speakerOnline{}
	BusSpeakerOffline       = speakerOffline{}
	BusSpeakerReonline      = speakerReonline{}
	BusSpeakerFormatChanged = speakerFormatChanged{}
)

type getSpeakers struct{}

func (getSpeakers) Dispatch(l *[]*Speaker) error {
	return bus.Dispatch("get speakers", l)
}
func (getSpeakers) Register(c func(l *[]*Speaker) error) *bus.HandlerData {
	return bus.Register("get speakers", func(_ any, a ...any) error {
		return c(a[0].(*[]*Speaker))
	})
}

type getSpeaker struct{}

func (getSpeaker) Dispatch(l *Speaker) error {
	return bus.Dispatch("get speaker", l)
}
func (getSpeaker) Register(c func(l *Speaker) error) *bus.HandlerData {
	return bus.Register("get speaker", func(_ any, a ...any) error {
		return c(a[0].(*Speaker))
	})
}

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

func (speakerEdited) Dispatch(sp *Speaker, field string, value, oldValue any) error {
	return bus.DispatchObj(sp, "speaker edited", field, value)
}
func (speakerEdited) Register(c func(sp *Speaker, field string, value, oldValue any) error) *bus.HandlerData {
	return bus.Register("speaker edited", func(o any, a ...any) error {
		return c(o.(*Speaker), a[0].(string), a[1], a[2])
	})
}

type speakerSaved struct{}

func (speakerSaved) Dispatch(l *Speaker) error {
	return bus.DispatchObj(l, "speaker saved")
}
func (speakerSaved) Register(c func(l *Speaker) error) *bus.HandlerData {
	return bus.Register("speaker saved", func(o any, _ ...any) error {
		return c(o.(*Speaker))
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
		return c(o.(*Speaker))
	})
}

type speakerOffline struct{}

func (speakerOffline) Dispatch(sp *Speaker) error {
	return bus.DispatchObj(sp, "speaker offline")
}
func (speakerOffline) Register(c func(sp *Speaker) error) *bus.HandlerData {
	return bus.Register("speaker offline", func(o any, a ...any) error {
		return c(o.(*Speaker))
	})
}

type speakerOnline struct{}

func (speakerOnline) Dispatch(sp *Speaker) error {
	return bus.DispatchObj(sp, "speaker online")
}
func (speakerOnline) Register(c func(sp *Speaker) error) *bus.HandlerData {
	return bus.Register("speaker online", func(o any, a ...any) error {
		return c(o.(*Speaker))
	})
}

type speakerReonline struct{}

func (speakerReonline) Dispatch(sp *Speaker) error {
	return bus.DispatchObj(sp, "speaker reonline")
}
func (speakerReonline) Register(c func(sp *Speaker) error) *bus.HandlerData {
	return bus.Register("speaker reonline", func(o any, a ...any) error {
		return c(o.(*Speaker))
	})
}

type speakerFormatChanged struct{}

func (speakerFormatChanged) Dispatch(sp *Speaker, format, oldFormat audio.Sample) error {
	return bus.DispatchObj(sp, "speaker format changed")
}
func (speakerFormatChanged) Register(c func(sp *Speaker, format, oldFormat audio.Sample) error) *bus.HandlerData {
	return bus.Register("speaker format changed", func(o any, a ...any) error {
		return c(o.(*Speaker), a[0].(audio.Sample), a[1].(audio.Sample))
	})
}
