package speaker

import "github.com/zwcway/castserver-go/common/bus"

var (
	BusGetLines    = getLines{}
	BusGetLine     = getLine{}
	BusLineCreated = lineCreated{}
	BusLineDeleted = lineDeleted{}

	BusLineOutputChanged = lineOutputChanged{}
)

type getLines struct{}

func (getLines) Dispatch(l *[]*Line) error {
	return bus.Dispatch("get lines", l)
}
func (getLines) Register(c func(l *[]*Line) error) *bus.HandlerData {
	return bus.Register("get lines", func(a ...any) error {
		return c(a[0].(*[]*Line))
	})
}

type getLine struct{}

func (getLine) Dispatch(l *Line) error {
	return bus.Dispatch("get line", l)
}
func (getLine) Register(c func(l *Line) error) *bus.HandlerData {
	return bus.Register("get line", func(a ...any) error {
		return c(a[0].(*Line))
	})
}

type lineCreated struct{}

func (lineCreated) Dispatch(l *Line) error {
	return bus.Dispatch("line created", l)
}
func (lineCreated) Register(c func(l *Line) error) *bus.HandlerData {
	return bus.Register("line created", func(a ...any) error {
		return c(a[0].(*Line))
	})
}

type lineDeleted struct{}

func (lineDeleted) Dispatch(src *Line, dst *Line) error {
	return bus.Dispatch("line deleted", src, dst)
}
func (lineDeleted) Register(c func(src *Line, dst *Line) error) *bus.HandlerData {
	return bus.Register("line deleted", func(a ...any) error {
		return c(a[0].(*Line), a[1].(*Line))
	})
}

type lineOutputChanged struct{}

func (lineOutputChanged) Dispatch(l *Line) error {
	return bus.Dispatch("line deleted", l)
}
func (lineOutputChanged) Register(l *Line, c func(l *Line) error) *bus.HandlerData {
	return bus.RegisterObj(l, "line deleted", func(a ...any) error {
		return c(a[0].(*Line))
	})
}
