// Package rpc
// @Title  title
// @Description  desc
// @Author  pc  2024/11/5
// @Update  pc  2024/11/5
package rpc

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/errdef"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/msgenvelope"
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

	methodMap map[string]*MethodInfo
	isPublic  bool // 是否是公开服务(有rpc调用的服务)
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Init(rpcHandler inf.IRpcHandler) {
	h.IRpcHandler = rpcHandler
	h.methodMap = make(map[string]*MethodInfo)

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
	h.methodMap[name] = methodInfo
	return nil
}

func (h *Handler) HandleRequest(envelope inf.IEnvelope) {
	defer func() {
		if r := recover(); r != nil {
			log.SysLogger.Errorf("service[%s] handle message panic: %v\n trace:%s", h.GetName(), r, debug.Stack())
		}
	}()

	var (
		params  []reflect.Value
		results []reflect.Value
	)
	methodInfo, ok := h.methodMap[envelope.GetMethod()]
	if !ok {
		envelope.SetError(errdef.MethodNotFound)
		goto DoResponse
	}

	params = append(params, reflect.ValueOf(h.GetRpcHandler()))
	if len(methodInfo.In) > 0 {
		// 有输入参数
		params = append(params, reflect.ValueOf(envelope.GetRequest()))
	}

	results = methodInfo.Method.Func.Call(params)
	if len(results) != len(methodInfo.Out) {
		// 这里应该不会触发,因为参数检查的时候已经做过了
		log.SysLogger.Errorf("method[%s] return value count not match", envelope.GetMethod())
		envelope.SetError(errdef.OutputParamNotMatch)
		goto DoResponse
	}

	if len(results) == 0 {
		// 没有返回值
		goto DoResponse
	}

	// 解析返回
	for i, t := range methodInfo.Out {
		result := results[i]
		if t.Kind() == reflect.Ptr ||
			t.Kind() == reflect.Interface ||
			t.Kind() == reflect.Func ||
			t.Kind() == reflect.Map ||
			t.Kind() == reflect.Slice ||
			t.Kind() == reflect.Chan {
			if t.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				if err, ok := result.Interface().(error); ok && err != nil {
					envelope.SetError(err)
					goto DoResponse
				}
			} else {
				if result.IsNil() {
					envelope.SetResponse(nil)
				} else {
					envelope.SetResponse(result.Interface())
				}
			}
		}
	}

DoResponse:
	if envelope.NeedResponse() {
		// 需要回复
		envelope.SetReply()      // 这是回复
		envelope.SetRequest(nil) // 清除请求数据

		// 发送回复信息
		if err := envelope.GetSenderClient().SendResponse(envelope); err != nil {
			log.SysLogger.Errorf("service[%s] send response failed: %v", h.GetName(), err)
			msgenvelope.ReleaseMsgEnvelope(envelope)
		}
	} else {
		// 不需要回复,释放资源
		msgenvelope.ReleaseMsgEnvelope(envelope)
	}
}

func (h *Handler) HandleResponse(envelope inf.IEnvelope) {
	defer func() {
		if r := recover(); r != nil {
			log.SysLogger.Errorf("service[%s] handle message panic: %v\n trace:%s", h.GetName(), r, debug.Stack())
		}
	}()

	// 执行回调
	envelope.RunCompletions()

	// 释放资源
	msgenvelope.ReleaseMsgEnvelope(envelope)
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
