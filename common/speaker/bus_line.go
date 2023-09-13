package speaker

import (
	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/stream"
)

// 声明事件参数列表
var (
	BusGetLines    = getLines{}
	BusGetLine     = getLine{}
	BusLineCreated = lineCreated{}
	BusLineEdited  = lineEdited{}
	BusLineDeleted = lineDeleted{}
	BusLineRefresh = lineRefresh{}

	BusLineInputChanged    = lineInputChanged{}
	BusLineOutputChanged   = lineOutputChanged{}
	BusLineNameChanged     = lineNameChanged{}
	BusLineVolumeChanged   = lineVolumeChanged{}
	BusLineSpeakerAppended = lineSpeakerAppended{}
	BusLineSpeakerRemoved  = lineSpeakerRemoved{}
)

type getLines struct{}

func (getLines) Dispatch(l *[]*Line) error {
	return bus.Dispatch("get lines", l)
}
func (getLines) Register(c func(l *[]*Line) error) *bus.HandlerData {
	return bus.Register("get lines", func(o any, a ...any) error {
		return c(a[0].(*[]*Line))
	})
}

type getLine struct{}

func (getLine) Dispatch(l *Line) error {
	return bus.Dispatch("get line", l)
}
func (getLine) Register(c func(l *Line) error) *bus.HandlerData {
	return bus.Register("get line", func(o any, a ...any) error {
		return c(a[0].(*Line))
	})
}

type lineCreated struct{}

func (lineCreated) Dispatch(l *Line) error {
	return bus.Dispatch("line created", l)
}
func (lineCreated) Register(c func(l *Line) error) *bus.HandlerData {
	return bus.Register("line created", func(o any, a ...any) error {
		return c(a[0].(*Line))
	})
}

type lineEdited struct{}

func (lineEdited) Dispatch(l *Line, args ...any) error {
	return bus.DispatchObj(l, "line edited", args...)
}
func (lineEdited) Register(c func(l *Line, args ...any) error) *bus.HandlerData {
	return bus.Register("line edited", func(o any, a ...any) error {
		return c(o.(*Line), a...)
	})
}

type lineDeleted struct{}

func (lineDeleted) Dispatch(src *Line, dst *Line) error {
	return bus.DispatchObj(src, "line deleted", dst)
}
func (lineDeleted) Register(c func(src *Line, dst *Line) error) *bus.HandlerData {
	return bus.Register("line deleted", func(o any, a ...any) error {
		return c(o.(*Line), a[0].(*Line))
	})
}

type lineInputChanged struct{}

func (lineInputChanged) Dispatch(l *Line, newSS stream.SourceStreamer) error {
	return bus.DispatchObj(l, "line output changed", newSS)
}
func (lineInputChanged) Register(c func(l *Line, newSS stream.SourceStreamer) error) *bus.HandlerData {
	return bus.Register("line output changed", func(o any, a ...any) error {
		return c(o.(*Line), a[0].(stream.SourceStreamer))
	})
}

type lineOutputChanged struct{}

func (lineOutputChanged) Dispatch(l *Line, oldFormat *audio.Format) error {
	return bus.DispatchObj(l, "line output changed", oldFormat)
}
func (lineOutputChanged) Register(c func(l *Line, oldFormat *audio.Format) error) *bus.HandlerData {
	return bus.Register("line output changed", func(o any, a ...any) error {
		return c(o.(*Line), a[0].(*audio.Format))
	})
}

type lineNameChanged struct{}

func (lineNameChanged) Dispatch(l *Line, oldName *string) error {
	return bus.DispatchObj(l, "line name changed", oldName)
}
func (lineNameChanged) Register(c func(l *Line, oldName *string) error) *bus.HandlerData {
	return bus.Register("line name changed", func(o any, a ...any) error {
		return c(o.(*Line), a[0].(*string))
	})
}

type lineVolumeChanged struct{}

func (lineVolumeChanged) Dispatch(l *Line, oldVol float64) error {
	return bus.DispatchObj(l, "line volume changed", oldVol)
}
func (lineVolumeChanged) Register(c func(l *Line, oldVol float64) error) *bus.HandlerData {
	return bus.Register("line volume changed", func(o any, a ...any) error {
		return c(o.(*Line), a[0].(float64))
	})
}

type lineSpeakerAppended struct{}

func (lineSpeakerAppended) Dispatch(l *Line, sp *Speaker) error {
	return bus.DispatchObj(l, "line speaker appended", sp)
}
func (lineSpeakerAppended) Register(c func(l *Line, sp *Speaker) error) *bus.HandlerData {
	return bus.Register("line speaker appended", func(o any, a ...any) error {
		return c(o.(*Line), a[0].(*Speaker))
	})
}

type lineSpeakerRemoved struct{}

func (lineSpeakerRemoved) Dispatch(l *Line, sp *Speaker) error {
	return bus.DispatchObj(l, "line speaker removed", sp)
}
func (lineSpeakerRemoved) Register(c func(l *Line, sp *Speaker) error) *bus.HandlerData {
	return bus.Register("line speaker removed", func(o any, a ...any) error {
		return c(o.(*Line), a[0].(*Speaker))
	})
}

type lineRefresh struct{}

func (lineRefresh) Dispatch(l *Line) error {
	return bus.DispatchObj(l, "line refresh")
}
func (lineRefresh) Register(c func(l *Line) error) *bus.HandlerData {
	return bus.Register("line refresh", func(o any, a ...any) error {
		return c(o.(*Line))
	})
}
