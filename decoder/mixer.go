package decoder

// 音频混合器
type Mixer struct {
	streamers []Streamer
}

func (m *Mixer) Len() int {
	return len(m.streamers)
}

func (m *Mixer) Add(s ...Streamer) {
	m.streamers = append(m.streamers, s...)
}

func (m *Mixer) Clear() {
	m.streamers = m.streamers[:0]
}

func (m *Mixer) Stream(samples *Samples) (int, bool) {
	var tmp = NewSamples(512, samples.Format)

	j := 0
	for j < samples.Size {
		mixed := 0
		for si := 0; si < len(m.streamers); si++ {
			// 混合音频流
			m.streamers[si].Stream(tmp)
			for ch := 0; ch < tmp.Format.Layout.Count; ch++ {
				for i := 0; i < tmp.LastSize; i++ {
					samples.Buffer[ch][j+i] += tmp.Buffer[ch][i]
				}
			}
			if mixed < tmp.LastSize {
				mixed = tmp.LastSize
			}
			if tmp.LastErr != nil {
				// 移除出有问题的音频流
				sj := len(m.streamers) - 1
				m.streamers[si], m.streamers[sj] = m.streamers[sj], m.streamers[si]
				m.streamers = m.streamers[:sj]
				si--
			}
		}
		j += mixed
	}

	return j, true
}
