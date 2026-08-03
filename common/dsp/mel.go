package dsp

import "math"

// LinearToMel 将线性频率转换为梅尔频率
func LinearToMel(freq float64) float64 {
	return 2595.0 * math.Log10(1.0+freq/700.0)
}

// MelToLinear 将梅尔频率转换回线性频率
func MelToLinear(mel float64) float64 {
	return 700.0 * (math.Pow(10.0, mel/2595.0) - 1.0)
}

// MelFilterElement 存储单个 FFT bin 对应的滤波器权重
type MelFilterElement struct {
	BinIndex int     // FFT 幅值数组的索引
	Weight   float64 // 该索引对应的三角形滤波器权重 (0.0 ~ 1.0)
}

// MelFilterBank 梅尔滤波器组缓存结构体
type MelFilterBank struct {
	NumMelBins int
	// 使用二维切片，每个梅尔频段只记录【权重不为 0】的稀疏元素，极大优化内存与遍历耗时
	Filters     [][]MelFilterElement
	CenterFreqs []float64
}

// NewMelFilterBank 初始化梅尔滤波器组（第一部分：计算斜率并缓存）
// fftSize: FFT 幅值数组长度（如 1025）
// sampleRate: 采样率（如 16000）
// numMelBins: 梅尔频段数（如 40）
func NewMelFilterBank(fftSize int, sampleRate int, numMelBins int) *MelFilterBank {
	bank := &MelFilterBank{
		NumMelBins:  numMelBins,
		Filters:     make([][]MelFilterElement, numMelBins),
		CenterFreqs: make([]float64, numMelBins),
	}

	maxFreq := float64(sampleRate) / 2
	minMel := LinearToMel(0.0)
	maxMel := LinearToMel(maxFreq)

	// 计算梅尔刻度上的等间隔点
	melPoints := make([]float64, numMelBins+2)
	for i := range melPoints {
		mel := minMel + float64(i)*(maxMel-minMel)/float64(numMelBins+1)
		melPoints[i] = MelToLinear(mel)
	}

	// 将每个频段的中心频率（即 melPoints[1] 到 melPoints[numMelBins]）存入结构体
	for i := range numMelBins {
		bank.CenterFreqs[i] = melPoints[i+1] // i+1 即为当前滤波器的中心点
	}
	// 映射到 FFT bin 索引
	binPoints := make([]int, len(melPoints))
	for i := range binPoints {
		binPoints[i] = int(math.Floor((float64(fftSize) - 1) * melPoints[i] / maxFreq))
	}

	// 提前计算每个频段的稀疏矩阵权重
	for i := 0; i < numMelBins; i++ {
		startBin := binPoints[i]
		centerBin := binPoints[i+1]
		endBin := binPoints[i+2]

		var elements []MelFilterElement

		// 左半沿（上升斜率）
		if centerBin != startBin {
			for j := startBin; j < centerBin; j++ {
				if j < fftSize {
					weight := float64(j-startBin) / float64(centerBin-startBin)
					if weight > 0.0 { // 只缓存有效权重
						elements = append(elements, MelFilterElement{BinIndex: j, Weight: weight})
					}
				}
			}
		}

		// 右半沿（下降斜率）
		if endBin != centerBin {
			for j := centerBin; j <= endBin; j++ {
				if j < fftSize {
					weight := float64(endBin-j) / float64(endBin-centerBin)
					if weight > 0.0 { // 只缓存有效权重
						elements = append(elements, MelFilterElement{BinIndex: j, Weight: weight})
					}
				}
			}
		}

		bank.Filters[i] = elements
	}

	return bank
}

// Apply 最终将缓存的斜率应用到实时输入的幅值中（第二部分：高性能转换）
// outMelAmplitudes: 外部传入的切片（用于接收结果，避免函数内部频繁分配内存带来的 GC 开销）
// fftAmplitudes: 输入的 1025 长度线性频谱幅值
func (bank *MelFilterBank) Apply(outMelAmplitudes []float64, fftAmplitudes []float64) int {
	// 确保输出切片长度足够
	_ = outMelAmplitudes[bank.NumMelBins-1] // 边界检查消除优化（BCE）

	for i := 0; i < bank.NumMelBins; i++ {
		sum := 0.0
		elements := bank.Filters[i]

		for j := 0; j < len(elements); j++ {
			el := elements[j]
			sum += fftAmplitudes[el.BinIndex] * el.Weight
		}
		outMelAmplitudes[i] = sum
	}
	return bank.NumMelBins
}
