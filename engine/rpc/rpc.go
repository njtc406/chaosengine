// Package rpc
// @Title  title
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package rpc

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"reflect"
	"runtime/debug"
	"strings"
	"unicode"
	"unicode/utf8"
)

type MethodInfo struct {
	Method reflect.Method
	In     []reflect.Type
	Out    []reflect.Type
}

type Handler struct {
	inf.IRpcHandler

	internalMap map[string]*MethodInfo
	isPublic    bool // 是否是公开服务(有rpc调用的服务)
}

func (h *Handler) Init(rpcHandler inf.IRpcHandler) {
	h.IRpcHandler = rpcHandler
	h.internalMap = make(map[string]*MethodInfo)

	h.registerMethod()
}

func (h *Handler) registerMethod() {
	typ := reflect.TypeOf(h.IRpcHandler)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		err := h.suitableMethods(method)
		if err != nil {
			log.SysLogger.Panic(err)
		}
	}
}

func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

func (h *Handler) isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

func (h *Handler) suitableMethods(method reflect.Method) error {
	// 只有以API或者rpc开头的方法才注册
	if !strings.HasPrefix(method.Name, "API") && !strings.HasPrefix(method.Name, "API_") {
		if !strings.HasPrefix(method.Name, "RPC") && !strings.HasPrefix(method.Name, "RPC_") {
			// 不是API或者RPC开头的方法,直接返回
			return nil
		}

		if !h.isPublic {
			// 走到这说明有rpc方法
			h.isPublic = true
		}
	}

	var methodInfo *MethodInfo

	//1.最多两个参数(第一个是输入,第二个是输出)
	if method.Type.NumIn() > 2 {
		return fmt.Errorf("%s too many params", method.Name)
	}
	//2.判断参数类型,必须是其他地方可调用的
	var in []reflect.Type
	for i := 0; i < method.Type.NumIn(); i++ {
		if h.isExportedOrBuiltinType(method.Type.In(i)) == false {
			return fmt.Errorf("%s Unsupported parameter types", method.Name)
		}
		in = append(in, method.Type.In(i))
	}

	// 最多两个返回值,一个结果,一个错误
	var outs []reflect.Type
	if method.Type.NumOut() > 2 {
		return fmt.Errorf("%s too many return params", method.Name)
	}

	for i := 0; i < method.Type.NumOut(); i++ {
		outs = append(outs, method.Type.Out(i))
	}

	name := method.Name
	methodInfo.In = in // 这里实际上不需要,如果每次调用都用反射检查输入参数,那么性能会降低
	methodInfo.Method = method
	methodInfo.Out = outs
	h.internalMap[name] = methodInfo
	return nil
}

func (h *Handler) Handle(envelope *actor.MsgEnvelope) {
	defer func() {
		if r := recover(); r != nil {
			log.SysLogger.Errorf("service[%s] handle message panic: %v\n trace:%s", h.GetName(), r, debug.Stack())
		}
	}()

}

func (h *Handler) GetName() string {
	return h.IRpcHandler.GetName()
}

func (h *Handler) GetPID() *actor.PID {
	return h.IRpcHandler.GetPID()
}

func (h *Handler) GetRpcHandler() inf.IRpcHandler {
	return h.IRpcHandler
}

func (h *Handler) IsPrivate() bool {
	return !h.isPublic
}

func (h *Handler) IsClosed() bool {
	return h.IRpcHandler.IsClosed()
}
