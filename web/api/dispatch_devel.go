//go:build debug
// +build debug

package api

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type jsonReqMessage struct {
	body []byte
}

func (j *jsonReqMessage) RequestId() string { return "" }
func (j *jsonReqMessage) Command() string   { return "" }
func (j *jsonReqMessage) Unmarshal(v any) error {
	return json.Unmarshal(j.body, v)
}

type jsonResp struct {
	Ret any
	Err string
}

func ApiDispatchDevel(ctx *fasthttp.RequestCtx) bool {
	path := string(ctx.Path())
	if len(path) < 5 || path[:5] != "/api/" {
		return false
	}
	r, ok := apiRouterList[path[5:]]
	if !ok {
		return false
	}
	req := jsonReqMessage{
		body: ctx.Request.Body(),
	}
	ret, err := r.cb(nil, &req, log)
	resp := jsonResp{
		Ret: ret,
	}
	if err != nil {
		resp.Err = err.Error()
	}
	b, err := json.Marshal(resp)
	if err != nil {
		ctx.WriteString(err.Error())
	} else {
		ctx.Write(b)
	}

	return true
}
