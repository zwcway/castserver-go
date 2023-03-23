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
			samples.Reset()

			mixer.Stream(samples)

			if slient {
				goto __slient__
			}

			r.resample.Stream(samples)

			if samples.LastNbSamples == 0 {
				// err = samples.LastErr
				goto __slient__
			}

			r.bufSize = samples.LastSamplesSize()
			r.bufPos = 0

		}

		for ; r.bufPos < r.bufSize && n < len(p); r.bufPos += 2 {
			for ch = 0; ch < samples.Format.Layout.Count; ch++ {
				p[n+0] = samples.RawData[ch][r.bufPos+0]
				p[n+1] = samples.RawData[ch][r.bufPos+1]
				n += 2
			}
		}
	}

	// err = samples.LastErr

	return

__slient__:
	zeroBuf(p)
	n = len(p)
	return

}

func zeroBuf(p []byte) {
	for i := 0; i < len(p); i++ {
		p[i] = 0
	}
}
