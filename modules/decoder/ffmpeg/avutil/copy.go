package avutil

import (
	"unsafe"

	"github.com/zwcway/castserver-go/common/stream"
	"github.com/zwcway/castserver-go/common/utils"
)

// 复制GO内存至C内存
//
// dst 是 ffmpeg 中 planar 格式数据，类型 (uint8_t**)，buf[0]指向第一个样本
//
// dstChannels 声道数量
//
// dstLineSize dst 总长度
//
// @return 每声道已复制的字节数
func CopyToCBytes(dst unsafe.Pointer, src *stream.Samples, dstChannels int, dstLineSize int) int {
	var (
		dstS   []unsafe.Pointer
		copied = 0
		c      = 0
		dstCh  []byte
	)

	dstS = utils.MakeSlice[unsafe.Pointer](dst, dstChannels)
	for c = range dstChannels {
		dstCh = utils.MakeSlice[byte](dstS[c], dstLineSize)
		copied = copy(dstCh, src.RawData[c])
	}

	return copied
}

// CopyFromCBytes 复制C内存至GO内存。复制之前必须保证format一致
//
// src 是 ffmpeg 中 planar 格式数据，类型 (uint8_t**)，buf[0]指向第一个样本
//
// srcChannels src 的声道数量
//
// srcLineSamples src 的每声道总样本数
//
// dstSampleOffset 每声道复制时 dst 偏移的样本数
//
// srcSamplesOffset 每声道复制时 src 偏移的样本数
//
// @return 每声道已复制的字节数
func CopyFromCBytes(dst *stream.Samples, src unsafe.Pointer, srcChannels, srcLineSamples, dstSampleOffset, srcSamplesOffset int) int {
	var (
		srcS           []unsafe.Pointer
		srcCh          []byte
		copied         = 0
		c              = 0
		srcBytesOffset = dst.Format.SamplesSize(srcSamplesOffset)
		srcLineSize    = dst.Format.SamplesSize(srcLineSamples)
		dstBytesOffset = dst.Format.SamplesSize(dstSampleOffset)
	)
	srcS = utils.MakeSlice[unsafe.Pointer](src, srcChannels)
	for c = range srcChannels {
		srcCh = utils.MakeSlice[byte](srcS[c], srcLineSize)
		copied = copy(dst.RawData[c][dstBytesOffset:], srcCh[srcBytesOffset:])
	}

	copied = dst.Format.SamplesCount(copied)

	return copied
}
