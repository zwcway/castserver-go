package element

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/zwcway/castserver-go/common/audio"
	"github.com/zwcway/castserver-go/common/bus"
	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/common/utils"
)

// TestSpectrum_Concurrent_Timers 测试 Stream 和 Spectrum 函数在两个定时器中运行
// 主要验证：1) 锁的正确使用 2) FFT 变换是否正确
func TestSpectrum_Concurrent_Timers(t *testing.T) {
	bus.Init(utils.NewEmptyContext())

	s := NewSpectrum().(*Spectrum)
	s.On()
	s.SetLogAxis(false)
	defer s.Close()

	format := audio.DefaultFormat()
	rate := float64(format.Rate.ToInt())

	// 频率分辨率 = rate / n = 44100 / 2048 ≈ 21.5 Hz
	freqResolution := rate / float64(s.n)
	t.Logf("Frequency resolution: %.2f Hz per bin", freqResolution)
	t.Logf("Log axis enabled: %v (non-linear frequency mapping)", s.LogAxis())

	// 测试用的混合频率
	testFreqs := []float64{220.0, 440.0, 880.0, 2000.0, 8000.0, 20000.0}

	// 生成混合波形数据
	generateMixedWave := func(n int, tStart float64) []float64 {
		data := make([]float64, n)
		for i := 0; i < n; i++ {
			t := tStart + float64(i)
			var v float64
			for _, freq := range testFreqs {
				v += 0.3 * math.Sin(2*math.Pi*freq*t/rate)
			}
			data[i] = v
		}
		return data
	}

	var (
		mu           sync.Mutex
		spectrumOK   bool
		peakValues   []float64
		spectrumLens []int
	)

	// Stream 定时器：写入混合波形 (2ms 间隔)
	streamTicker := time.NewTicker(time.Millisecond * 2)
	defer streamTicker.Stop()

	// Spectrum 定时器：读取并验证 FFT 结果 (5ms 间隔)
	spectrumTicker := time.NewTicker(time.Millisecond * 5)
	defer spectrumTicker.Stop()

	var wg sync.WaitGroup
	wg.Add(2)

	tStart := time.Now()
	sampleCounter := 0
	milliSecond := time.Duration(300 * time.Millisecond)
	iteration := 0

	// 定时器 1：Stream 写入数据
	go func() {
		defer wg.Done()
		for range streamTicker.C {
			if time.Since(tStart) > milliSecond {
				return
			}
			samples := stream.NewSamples(s.n, format)
			waveData := generateMixedWave(samples.RequestNbSamples, float64(sampleCounter))
			copy(samples.Data[0], waveData)
			samples.LastNbSamples = samples.RequestNbSamples
			sampleCounter += samples.RequestNbSamples
			s.Stream(samples)
		}
	}()

	// 定时器 2：Spectrum 读取数据并验证
	go func() {
		defer wg.Done()
		for range spectrumTicker.C {
			if time.Since(tStart) > milliSecond {
				return
			}
			iteration++

			spectrum := s.Spectrum()

			// 找到频谱中的峰值
			var maxVal float64
			for _, v := range spectrum {
				if v > maxVal {
					maxVal = v
				}
			}

			mu.Lock()
			spectrumLens = append(spectrumLens, len(spectrum))
			peakValues = append(peakValues, maxVal)
			mu.Unlock()

			// t.Logf("Iteration %d: spectrum len=%d, peak=%f", iteration, len(spectrum), maxVal)

			validBins := make([]int, 0)
			for i, v := range spectrum {
				// 只考虑大于峰值一半的频谱数据
				if v < maxVal*0.5 {
					continue
				}
				validBins = append(validBins, i)
			}

			notFoundFreqs := make([]int, 0)
			for j, freq := range testFreqs {
				for _, i := range validBins {
					validFreq := float64(i) * freqResolution
					// 检查频率是否接近预期的测试频率
					if math.Abs(validFreq-freq) > freqResolution {
						notFoundFreqs = append(notFoundFreqs, j)
					}
				}
			}
			if iteration == 1 {
				for _, i := range validBins {
					v := spectrum[i]
					t.Logf("spectrum[%d] %6.0f Hz -> %.1f", i, float64(i)*freqResolution, v)
				}
			}
			if s.logAxis && len(notFoundFreqs) > 0 {
				for _, v := range s.melFilter.CenterFreqs {
					t.Logf("      freqs %.0f   %5d:%d", v, 0, s.n/2)
				}
				// t.Logf("log bins freqs len=%d %v", len(s.binsFreqs), freqs)
			}

			// 如果有测试频率没有在频谱中找到，则打印频谱信息
			for _, j := range notFoundFreqs {
				freq := testFreqs[j]
				t.Errorf("Frequency %f Hz not found in spectrum", freq)
				return
			}

		}
	}()

	wg.Wait()

	// 验证结果
	mu.Lock()
	defer mu.Unlock()

	// 检查是否获取到了有效的频谱数据
	if len(peakValues) == 0 {
		t.Fatal("No spectrum data received during test")
	}

	// 验证频谱长度一致且非零
	for i, l := range spectrumLens {
		if l == 0 {
			t.Errorf("Spectrum length at iteration %d is 0", i)
		}
	}

	// 验证峰值是否合理（对于有输入信号的情况，峰值应该 > 0）
	for i, v := range peakValues {
		if v <= 0 {
			t.Errorf("Peak value at iteration %d is %f, expected > 0 for signal input", i, v)
		} else {
			spectrumOK = true
		}
	}

	if spectrumOK {
		t.Logf("✓ Concurrent test passed: Stream and Spectrum functions work correctly with timers")
		t.Logf("  Captured %d spectrum samples, all with valid peak values", len(peakValues))
	}
}

func TestSpectrum_initMelFilter(t *testing.T) {
	r := NewSpectrum().(*Spectrum)
	r.bins = 10
	r.initMelFilter()

	for _, v := range r.melFilter.CenterFreqs {
		t.Logf("      freqs %.0f   %5d:%d", v, 0, r.n/2)
	}

}
