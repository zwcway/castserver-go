package utils

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
)

type ctxValue int

const (
	valueSignalKey ctxValue = iota + 1
	valueLoggerKey
)

type Context interface {
	context.Context
	Signal() chan os.Signal
	Logger(tag string) *zap.Logger
}

type cContext struct {
	c context.Context
}

func (c *cContext) Deadline() (time.Time, bool) { return c.c.Deadline() }
func (c *cContext) Err() error                  { return c.c.Err() }
func (c *cContext) Value(key any) any           { return c.c.Value(key) }
func (c *cContext) Done() <-chan struct{}       { return c.c.Done() }

func (c *cContext) Signal() chan os.Signal {
	return c.c.Value(valueSignalKey).(chan os.Signal)
}
func (c *cContext) logger() *zap.Logger {
	return c.c.Value(valueLoggerKey).(*zap.Logger)
}
func (c *cContext) Logger(tag string) *zap.Logger {
	return c.logger().With(zap.String("tag", tag))
}

func (c *cContext) WithSignal(sig chan os.Signal) *cContext {
	c.c = context.WithValue(c.c, valueSignalKey, sig)
	return c
}
func (c *cContext) WithLogger(log *zap.Logger) *cContext {
	c.c = context.WithValue(c.c, valueLoggerKey, log)
	return c
}

func (c *cContext) WithCancel() (*cContext, context.CancelFunc) {
	var cancel context.CancelFunc
	c.c, cancel = context.WithCancel(c.c)
	return c, cancel
}

func NewContext() *cContext {
	cc := cContext{context.Background()}
	return &cc
}
