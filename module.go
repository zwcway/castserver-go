package main

import (
	utils "github.com/zwcway/castserver-go/common/utils"
)

type Module interface {
	Init(ctx utils.Context) error
	DeInit()
}
