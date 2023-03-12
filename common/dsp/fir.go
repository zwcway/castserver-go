package dsp

type FIR struct {
	state []float64
	coefs []float64
}

func NewFIR(coefs []float64) *FIR {
	c := make([]float64, len(coefs))
	s := make([]float64, len(coefs))
	for i, v := range coefs {
		c[i] = v
	}
	return &FIR{coefs: c, state: s}
}

func (f *FIR) Step(in float64) float64 {
	result := 0.0
	for i := 0; i < len(f.state)-1; i++ {
		f.state[i] = f.state[i+1]
		result += f.state[i+1] * f.coefs[i]
	}

	f.state[len(f.state)-1] = in
	result += f.coefs[len(f.state)-1] * in
	return result
}

type IIR struct {
	A     []float64
	B     []float64
	state []float64
}

func NewIIR(A, B []float64) *IIR {
	l := 0
	if len(A)+1 > len(B) {
		l = len(A) + 1
	} else {
		l = len(B)
	}

	a := make([]float64, l-1)
	b := make([]float64, l)

	for i, v := range A {
		a[i] = v
	}

	for i, v := range B {
		b[i] = v
	}

	s := make([]float64, l)
	return &IIR{A: a, B: b, state: s}
}

func (f *IIR) Step(in float64) float64 {
	result := 0.0
	for i := 0; i < len(f.state)-1; i++ {
		f.state[i+1] = f.state[i]
		result += -1 * f.state[i] * f.A[i]
	}
	f.state[0] = result
	result = 0
	for i, v := range f.B {
		result += f.state[i] * v
	}
	return result
}
