package dsp

import (
	"math"
)

const Pi float64 = math.Pi

// Robert Bristow-Johnson's audio EQ cookbook
type Filter struct {
	t FilterType

	FilterParams

	in1  float64
	in2  float64
	out1 float64
	out2 float64

	a0 float64
	a1 float64
	a2 float64
	b0 float64
	b1 float64
	b2 float64
}

func (f *Filter) Process(input float64) float64 {
	output := (f.b0/f.a0)*input +
		(f.b1/f.a0)*f.in1 +
		(f.b2/f.a0)*f.in2 -
		(f.a1/f.a0)*f.out1 -
		(f.a2/f.a0)*f.out2

	f.in2 = f.in1
	f.in1 = input

	f.out2 = f.out1
	f.out1 = output

	return output
}

func (e *Filter) Init(rate int) {
	switch e.t {
	case LowPassFilter:
		e.initLowPass(rate)
	case HighPassFilter:
		e.initHighPass(rate)
	case PeakingFilter:
		e.initPeaking(rate)
	}
}

func (e *Filter) initLowPass(rate int) {
	q := e.Q
	w0 := 2.0 * Pi * float64(e.Frequency) / float64(rate)
	alpha := math.Sin(w0) / (2.0 * q)

	e.a0 = 1.0 + alpha
	e.a1 = -2.0 * math.Cos(w0)
	e.a2 = 1.0 - alpha
	e.b0 = (1.0 - math.Cos(w0)) / 2.0
	e.b1 = 1.0 - math.Cos(w0)
	e.b2 = (1.0 - math.Cos(w0)) / 2.0
}

func (e *Filter) initHighPass(rate int) {
	q := e.Q
	w0 := 2.0 * Pi * float64(e.Frequency) / float64(rate)
	alpha := math.Sin(w0) / (2.0 * q)

	e.a0 = 1.0 + alpha
	e.a1 = -2.0 * math.Cos(w0)
	e.a2 = 1.0 - alpha
	e.b0 = (1.0 + math.Cos(w0)) / 2.0
	e.b1 = -1.0 * (1.0 + math.Cos(w0))
	e.b2 = (1.0 + math.Cos(w0)) / 2.0
}

func (e *Filter) initPeaking(rate int) {
	width := 0.5
	w0 := 2.0 * Pi * float64(e.Frequency) / float64(rate)
	alpha := math.Sin(w0) * math.Sinh(math.Log(2.0)/2.0*width*w0/math.Sin(w0))
	a := math.Pow(10.0, (e.Gain / 40.0))

	e.a0 = 1.0 + alpha/a
	e.a1 = -2.0 * math.Cos(w0)
	e.a2 = 1.0 - alpha/a
	e.b0 = 1.0 + alpha*a
	e.b1 = -2.0 * math.Cos(w0)
	e.b2 = 1.0 - alpha*a
}

func NewFilter(eq FilterParams, rate int) *Filter {
	f := &Filter{FilterParams: eq}
	f.Init(rate)
	return f
}
