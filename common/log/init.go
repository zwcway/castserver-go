package log

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
)

var log *Log
var encoderConfig = zapcore.EncoderConfig{
	TimeKey:        "ts",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	FunctionKey:    zapcore.OmitKey,
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    levelEncoder,
	EncodeName:     nameEncoder,
	EncodeTime:     timeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

func NewLogger(logFile string, daemon bool) (l Logger, close func(), err error) {
	if log != nil {
		l = log
		return
	}

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	topicDebugging := zapcore.AddSync(io.Discard)
	topicErrors := zapcore.AddSync(io.Discard)
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	// consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleEncoder := NewConsoleEncoder(encoderConfig)

	cores := []zapcore.Core{}

	if len(logFile) > 0 {
		// TODO 接收信号重新打开日志文件
		var (
			sink zapcore.WriteSyncer
		)
		sink, close, err = zap.Open(logFile)
		if err != nil {
			close()
			err = errors.Wrap(err, "open log file error")
			return
		}
		cores = append(cores,
			zapcore.NewCore(fileEncoder, sink, highPriority),
			zapcore.NewCore(fileEncoder, sink, lowPriority),
		)
	} else {
		cores = append(cores,
			zapcore.NewCore(fileEncoder, topicErrors, highPriority),
			zapcore.NewCore(fileEncoder, topicDebugging, lowPriority),
		)
	}

	if !daemon {
		cores = append(cores,
			zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
			zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
		)
	}

	log = &Log{l: zap.New(zapcore.NewTee(cores...))}

	l = log

	return
}

func NewMemroy() Logger {
	log = &Log{l: zap.NewNop()}
	return log
}

func NewDBLog(level Level) logger.Interface {
	return &dbLog{log: log.Name("database"), lv: level}
}
