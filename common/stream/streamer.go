package stream

import (
	"time"

	"github.com/zwcway/castserver-go/common/audio"
)

type Streamer interface {
	Stream(*Samples)
}

type StreamCloser interface {
	Streamer
	Close() error
}

type StreamSeekCloser interface {
	StreamCloser
	Len() int      // 总长度
	Position() int // 当前位置
	Seek(p time.Duration) error
}

type SourceStreamer interface {
	StreamCloser
	AudioFormat() audio.Format         // 获取输入的音频格式
	ChannelIndex() audio.ChannelIndex // 获取输入的声道布局
	SetOutFormat(audio.Format) error   // 设置音频输出格式
	IsPlaying() bool
	CanRemove() bool // 是否可以自动移除
}

type FormatChangedHandler func(stream SourceStreamer, inFormat, outFormat audio.Format)

type FileStreamer interface {
	SourceStreamer
	StreamSeekCloser
	OpenFile(string) error
	CurrentFile() string
	Duration() time.Duration      // 当前时长
	TotalDuration() time.Duration // 总时长
	SetPause(bool)                // 暂停解码
	IsPaused() bool               // 是否暂停
}

type ReceiverStreamer interface {
	SourceStreamer
}
