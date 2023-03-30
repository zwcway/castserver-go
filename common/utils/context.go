package utils

import (
	"context"
	"os"
	"time"

	"github.com/zwcway/castserver-go/common/lg"
)

type ctxValue int

const (
	valueSignalKey ctxValue = iota + 1
	valueLoggerKey
)

type Context interface {
	context.Context
	Signal() chan os.Signal
	Logger(tag string) lg.Logger
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
func (c *cContext) logger() lg.Logger {
	return c.c.Value(valueLoggerKey).(lg.Logger)
}
func (c *cContext) Logger(tag string) lg.Logger {
	return c.logger().Name(tag)
}

func (c *cContext) WithSignal(sig chan os.Signal) *cContext {
	c.c = context.WithValue(c.c, valueSignalKey, sig)
	return c
}
func (c *cContext) WithLogger(log lg.Logger) *cContext {
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

func NewEmptyContext() *cContext {
	cc := cContext{context.Background()}
	cc.WithLogger(lg.NewMemroy())
	return &cc
}
