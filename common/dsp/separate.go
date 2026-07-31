package dsp

import (
	"math"
)

func MakeChannelData(channelCnt int, bitSize int, mtu int) map[int][]byte {
	m := make(map[int][]byte, channelCnt)
	maxSizePerCh := int(math.Floor(float64(mtu) / float64(bitSize)))

	for i := 0; i < channelCnt; i++ {
		m[i] = make([]byte, maxSizePerCh)
	}
	return m
}

func SepChannel(chData map[int][]byte, data []byte, chs []int, bs int) {
	dataLen := len(data)
	chCnt := len(chs)
	lenPerCh := dataLen / (chCnt * bs)

	if chCnt == 0 || dataLen == 0 {
		return
	}
	if dataLen%(chCnt*bs) == 0 {
		return
	}
	if len(chData) < chCnt {
		return
	}

	di := 0
	switch bs {
	case 1:
		for i := 0; i < lenPerCh; i += bs {
			for _, c := range chs {
				chData[c][i] = data[di]
				di += bs
			}
		}
	case 2:
		for i := 0; i < lenPerCh; i += bs {
			for _, c := range chs {
				chData[c][i] = data[di]
				chData[c][i+1] = data[di+1]
				di += bs
			}
		}
	case 3:
		for i := 0; i < lenPerCh; i += bs {
			for _, c := range chs {
				chData[c][i] = data[di]
				chData[c][i+1] = data[di+1]
				chData[c][i+2] = data[di+2]
				di += bs
			}
		}
	case 4:
		for i := 0; i < lenPerCh; i += bs {
			for _, c := range chs {
				chData[c][i] = data[di]
				chData[c][i+1] = data[di+1]
				chData[c][i+2] = data[di+2]
				chData[c][i+3] = data[di+3]
				di += bs
			}
		}

	}
}
