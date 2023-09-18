package log

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

type Field struct {
	zap.Field
}

func String(k string, v string) Field          { return Field{zap.String(k, v)} }
func Int(k string, v int64) Field              { return Field{zap.Int64(k, v)} }
func Uint(k string, v uint64) Field            { return Field{zap.Uint64(k, v)} }
func Binary(k string, v []byte) Field          { return Field{zap.Binary(k, v)} }
func ByteString(k string, v []byte) Field      { return Field{zap.ByteString(k, v)} }
func Bool(k string, v bool) Field              { return Field{zap.Bool(k, v)} }
func Duration(k string, v time.Duration) Field { return Field{zap.Duration(k, v)} }
func Float32(k string, v float32) Field        { return Field{zap.Float32(k, v)} }
func Float64(k string, v float64) Field        { return Field{zap.Float64(k, v)} }
func Error(v error) Field                      { return Field{zap.Error(v)} }
func Time(k string, v time.Time) Field         { return Field{zap.Time(k, v)} }

func Any(k string, v any) Field {
	if s, ok := v.(fmt.Stringer); ok {
		return Field{zap.String(k, s.String())}
	}
	return Field{zap.Any(k, v)}
}

type Logger interface {
	Debug(msg string, feilds ...Field)
	Info(msg string, feilds ...Field)
	Warn(msg string, feilds ...Field)
	Error(msg string, feilds ...Field)
	Fatal(msg string, feilds ...Field)
	Panic(msg string, feilds ...Field)

	Name(string) Logger
}

type Log struct {
	l  *zap.Logger
	zf []zap.Field
}

func (l *Log) Debug(msg string, feilds ...Field) {
	l.l.Debug(msg, l.zfields(feilds...)...)
}

func (l *Log) Info(msg string, feilds ...Field) {
	l.l.Info(msg, l.zfields(feilds...)...)
}

func (l *Log) Warn(msg string, feilds ...Field) {
	l.l.Warn(msg, l.zfields(feilds...)...)
}

func (l *Log) Error(msg string, feilds ...Field) {
	l.l.Error(msg, l.zfields(feilds...)...)
}

func (l *Log) Fatal(msg string, feilds ...Field) {
	l.l.Fatal(msg, l.zfields(feilds...)...)
}

func (l *Log) Panic(msg string, feilds ...Field) {
	l.l.Panic(msg, l.zfields(feilds...)...)
}

func (l *Log) Name(f string) Logger {
	return &Log{
		l: l.l.Named(strings.ToUpper(f)),
	}
}

func (l *Log) zfields(fields ...Field) []zap.Field {
	size := len(fields)
	if len(l.zf) < size {
		l.zf = make([]zap.Field, size*2)
	}
	for i := 0; i < len(fields); i++ {
		l.zf[i] = fields[i].Field
	}
	return l.zf[:size]
}

type dbLog struct {
	log Logger
	lv  Level
}

func (l *dbLog) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *dbLog) Info(ctx context.Context, msg string, args ...any) {
	if l.lv < InfoLevel {
		return
	}
	l.log.Info(fmt.Sprintf(msg, args...))
}

func (l *dbLog) Warn(ctx context.Context, msg string, args ...any) {
	if l.lv < WarnLevel {
		return
	}
	l.log.Warn(fmt.Sprintf(msg, args...))
}

func (l *dbLog) Error(ctx context.Context, msg string, args ...any) {
	if l.lv < ErrorLevel {
		return
	}
	l.log.Error(fmt.Sprintf(msg, args...))
}

func (l *dbLog) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.lv < DebugLevel {
		return
	}
	sql, rows := fc()
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	l.log.Debug(msg, String("sql", sql), Int("rows", rows), Time("begin", begin))
}
