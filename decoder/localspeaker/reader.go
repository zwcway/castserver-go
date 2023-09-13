package localspeaker

import (
	"time"

	"github.com/zwcway/castserver-go/common/stream"
)

type reader struct {
	samples  *stream.Samples
	resample stream.ResampleElement
	bufPos   int
	bufSize  int
}

func (r *reader) Read(p []byte) (n int, err error) {
	var (
		// r.samples 有可能异步被改变，使用变更之前的对象
		samples = r.samples
		ch      int
		t       = time.Now()
	)
	defer func() {
		cost = time.Since(t)
	}()

	if samples == nil || mixer.Len() == 0 {
		goto __slient__
	}

	for n < len(p) {
		if r.bufPos >= r.bufSize {
			samples.ResetData()

			mixer.Stream(samples)

			if slient || samples.LastNbSamples == 0 {
				goto __slient__
			}

			if r.resample != nil {
				r.resample.Stream(samples)
			}

			if samples.LastNbSamples == 0 || samples.Format != format {
				// err = samples.LastErr
				goto __slient__
			}

			r.bufSize = samples.LastSamplesSize()
			r.bufPos = 0

		}

		for ; r.bufPos < r.bufSize-1 && n < len(p)-1; r.bufPos += 2 {
			for ch = 0; ch < int(samples.Format.Count); ch++ {
				p[n+0] = samples.RawData[ch][r.bufPos+0]
				p[n+1] = samples.RawData[ch][r.bufPos+1]
				n += 2
			}
		}
	}

	// err = samples.LastErr

	return

__slient__:
	zeroBuf(p, n)
	n = len(p)
	return

}

func zeroBuf(p []byte, offset int) {
	for i := offset; i < len(p); i++ {
		p[i] = 0
	}
}
