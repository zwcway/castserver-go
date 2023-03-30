package config

import "time"

type MilliDuration = time.Duration

type InvalidError struct {
	ck  *CfgKey
	err string
}

func (e *InvalidError) Error() string {
	return e.err
}

type EmptyError struct {
	key string
}

func (e *EmptyError) Error() string {
	return ""
}
