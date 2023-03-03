package dsp

import "math"

func LevelMeterUint8(data []byte, step int) float64 {
	var (
		frac float64
		sum  float64
		rms  float64
	)

	for i := 0; i < len(data); i += 1 * step {
		frac = float64(data[i]) * 1.0 / 128.0 // 0x80
		sum += frac * frac
	}
	rms = math.Sqrt(sum * 1.0 / float64(len(data)))
	rms = math.Max(0.0, rms)
	rms = math.Min(1.0, rms)

	return rms
}
func LevelMeterUint16(data []byte, step int) float64 {
	var (
		frac float64
		sum  float64
		rms  float64
	)

	for i := 0; i < len(data); i += 2 * step {
		frac = float64(int(data[i])+(int(data[i+1])<<8)) * 1.0 / 32768.0 // 0x8000
		sum += frac * frac
	}
	rms = math.Sqrt(sum * 1.0 / float64(len(data)))
	rms = math.Max(0.0, rms)
	rms = math.Min(1.0, rms)

	return rms
}
func LevelMeterUint24(data []byte, step int) float64 {
	var (
		frac float64
		sum  float64
		rms  float64
	)

	for i := 0; i < len(data); i += 3 * step {
		frac = float64(int(data[i])+(int(data[i+1])<<8)+(int(data[i+2])<<16)) * 1.0 / 8388608.0 // 0x800000
		sum += frac * frac
	}
	rms = math.Sqrt(sum * 1.0 / float64(len(data)))
	rms = math.Max(0.0, rms)
	rms = math.Min(1.0, rms)

	return rms
}
func LevelMeterUint32(data []byte, step int) float64 {
	var (
		frac float64
		sum  float64
		rms  float64
	)

	for i := 0; i < len(data); i += 4 * step {
		frac = float64(int(data[i])+(int(data[i+1])<<8)+(int(data[i+2])<<16)+(int(data[i+3])<<24)) * 1.0 / 2147483648.0 // 0x80000000
		sum += frac * frac
	}
	rms = math.Sqrt(sum * 1.0 / float64(len(data)))
	rms = math.Max(0.0, rms)
	rms = math.Min(1.0, rms)

	return rms
}

func LevelMeterFloat64(data []float64) float64 {
	var (
		sum float64
		rms float64
	)

	for i := 0; i < len(data); i++ {
		sum += data[i] * data[i]
	}

	rms = math.Sqrt(sum * 1.0 / float64(len(data)))
	rms = math.Max(0.0, rms)
	rms = math.Min(1.0, rms)

	return rms
}
