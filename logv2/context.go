package logv2

import (
	"context"
	"fmt"
	"github.com/rs/xid"
	"runtime"
	"strings"
)

const (
	funcKey ctxKey = "func"
	flowKey ctxKey = "flow"
	reqId   ctxKey = "req_id"
)

type ctxKey string

func setCtxFuncName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, funcKey, name)
}
func getCtxFunc(ctx context.Context) (funcName string) {
	if _, ok := ctx.Value(funcKey).(string); ok {
		funcName = ctx.Value(funcKey).(string)
	}
	return
}

func appendCtxFuncFlow(ctx context.Context, flow string) context.Context {
	if _, ok := ctx.Value(flowKey).(string); ok {
		ctx = context.WithValue(ctx, flowKey, fmt.Sprintf("%s-%s", ctx.Value(flowKey).(string), flow))
	} else {
		ctx = context.WithValue(ctx, flowKey, flow)
	}

	return ctx
}
func getCtxFlow(ctx context.Context) (flow string) {
	if _, ok := ctx.Value(flowKey).(string); ok {
		flow = ctx.Value(flowKey).(string)
	}
	return
}

func setCtxReqId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, reqId, id)
}
func getCtxReqId(ctx context.Context) (id string) {
	if _, ok := ctx.Value(reqId).(string); ok {
		id = ctx.Value(reqId).(string)
	}
	return
}

func getCallerFuncName() string {
	funcName := ""
	rc, _, _, ok := runtime.Caller(2)
	details := runtime.FuncForPC(rc)
	if ok && nil != details {
		fName := strings.Split(details.Name(), ".")
		if len(fName) > 0 {
			funcName = fName[len(fName)-1]
		}
	}

	return funcName
}

func InitFuncContext(ctx context.Context) context.Context {
	funcName := getCallerFuncName()
	ctx = setCtxFuncName(ctx, funcName)
	ctx = appendCtxFuncFlow(ctx, funcName)
	return ctx
}

func InitRequestContext(ctx context.Context) context.Context {
	ctx = setCtxReqId(ctx, xid.New().String())
	return ctx
}
