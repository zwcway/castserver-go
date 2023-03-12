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

type FileStreamer interface {
	StreamSeekCloser
	OpenFile(string) error
	CurrentFile() string
	AudioFormat() audio.Format            // 当前音频文件格式
	OutAudioFormat() audio.Format         // 音频输出格式
	SetOutAudioFormat(audio.Format) error // 设置音频输出格式
	Duration() time.Duration              // 当前时长
	TotalDuration() time.Duration         // 总时长
	Pause(bool)                           // 暂停解码
	IsPaused() bool                       // 是否暂停
}
