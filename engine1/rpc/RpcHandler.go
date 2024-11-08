// Package rpc
// @Title  服务rpc接口
// @Description  服务rpc接口
// @Author  yr  2024/7/19 上午10:42
// @Update  yr  2024/7/19 上午10:42
package rpc

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine1/actor"
	"github.com/njtc406/chaosengine/engine1/cluster/endpoints"
	"github.com/njtc406/chaosengine/engine1/define/inf"
	"github.com/njtc406/chaosengine/engine1/errdef/errcode"
	"github.com/njtc406/chaosengine/engine1/log"
	"github.com/njtc406/chaosengine/engine1/utils"
	"github.com/njtc406/chaosutil/chaoserrors"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	defaultTimeout = time.Second * 2
)

type MethodInfo struct {
	method reflect.Method
	in     []reflect.Type // 参数类型
	outs   []reflect.Type // 返回值类型
}

type RawCallback func(msg interface{})

var nilError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())

type Handler struct {
	inf.IChannel

	handler inf.IRpcHandler

	rawFunction     RawCallback // 客户端消息接收器
	mapFunctions    map[string]*MethodInfo
	mapRpcFunctions map[string]*MethodInfo
}

func (ch *Handler) InitHandler(handler inf.IRpcHandler, rpcChannel inf.IChannel) {
	ch.handler = handler
	ch.IChannel = rpcChannel
	ch.mapFunctions = make(map[string]*MethodInfo)
	ch.mapRpcFunctions = make(map[string]*MethodInfo)
	ch.registerHandler(handler)
}

func (ch *Handler) GetName() string {
	return ch.handler.GetName()
}

func (ch *Handler) registerHandler(rpcHandler inf.IRpcHandler) {
	typ := reflect.TypeOf(rpcHandler)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		err := ch.suitableMethods(method)
		if err != nil {
			log.SysLogger.Panic(err)
		}
		err = ch.suitableRpcMethods(method)
		if err != nil {
			log.SysLogger.Panic(err)
		}
	}
}

func (ch *Handler) RegisterRawHandler(callback RawCallback) {
	ch.rawFunction = callback
}

func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

func (ch *Handler) isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// suitableMethods 注册方法
func (ch *Handler) suitableMethods(method reflect.Method) error {
	// 只有以API开头的方法才注册
	if !strings.HasPrefix(method.Name, "API") && !strings.HasPrefix(method.Name, "API_") {
		return nil
	}

	var methodInfo MethodInfo
	//2.判断参数类型,必须是其他地方可调用的
	for i := 0; i < method.Type.NumIn(); i++ {
		if ch.isExportedOrBuiltinType(method.Type.In(i)) == false {
			return fmt.Errorf("%s Unsupported parameter types", method.Name)
		}
	}

	// 只关心返回值,因为返回值除了错误,其他都需要被放入response
	// 比如返回 int,int,error, 那么最终返回的是[]interface{int, int}, error
	// 如果直接从函数返回那里来判断返回,interface类型的error会返回nil, 而不是error, 所以在这里纪录一下返回值类型
	var outs []reflect.Type
	for i := 0; i < method.Type.NumOut(); i++ {
		outs = append(outs, method.Type.Out(i))
	}
	name := method.Name
	methodInfo.method = method
	methodInfo.outs = outs
	ch.mapFunctions[name] = &methodInfo
	return nil
}

func (ch *Handler) suitableRpcMethods(method reflect.Method) error {
	// 只有以RPC开头的方法才注册
	if !strings.HasPrefix(method.Name, "RPC") && !strings.HasPrefix(method.Name, "RPC_") {
		return nil
	}

	var methodInfo MethodInfo
	//2.判断参数类型,必须是其他地方可调用的

	var in []reflect.Type
	for i := 1; i < method.Type.NumIn(); i++ {
		if ch.isExportedOrBuiltinType(method.Type.In(i)) == false {
			return fmt.Errorf("%s Unsupported parameter types", method.Name)
		}
		in = append(in, method.Type.In(i))
	}

	var outs []reflect.Type
	for i := 0; i < method.Type.NumOut(); i++ {
		outs = append(outs, method.Type.Out(i))
	}
	name := method.Name
	methodInfo.method = method
	methodInfo.in = in
	methodInfo.outs = outs
	ch.mapRpcFunctions[name] = &methodInfo
	return nil
}

//func (ch *Handler) GetName() string {
//	return ch.handler.GetName()
//}

func (ch *Handler) GetPID() *actor.PID {
	return ch.handler.GetPID()
}

func (ch *Handler) GetRpcHandler() inf.IRpcHandler {
	return ch.handler
}

// sendMessage 发送消息(无回复)
func (ch *Handler) sendMessage(nodeUID, serviceID, serviceMethod string, isNode bool, args ...interface{}) chaoserrors.CError {
	serviceName, method, err := utils.SplitServiceMethod(serviceMethod)
	if err != nil {
		return chaoserrors.NewErrCode(errcode.ServiceMethodError, err)
	}

	// 获取客户端
	clients := endpoints.GetEndpointManager().GetServiceByID(nodeUID, serviceID, serviceName, false)
	if len(clients) == 0 {
		log.SysLogger.Errorf("can not find service method [%s] id [%s]", serviceMethod, serviceID)
		return chaoserrors.NewErrCode(errcode.ServerError, "can not find service", nil)
	} else if len(clients) > 1 {
		log.SysLogger.Errorf("can not call service method [%s] id [%s] more than 1 service", serviceMethod, serviceID)
		return chaoserrors.NewErrCode(errcode.ServerError, "call method more than 1 service", nil)
	}

	envelope := actor.NewMsgEnvelope()
	envelope.Method = method
	//envelope.Sender = ch.GetPID() // send不需要发送者
	envelope.Receiver = clients[0].GetPID()
	envelope.ReqID = clients[0].GenerateSeq()
	if isNode {
		if len(args) > 0 {
			envelope.SetArgs(args[0])
		}
	} else {
		envelope.SetArgs(args)
	}

	// 向客户端发送消息
	// (这里的envelope会在两个地方回收,如果是本地调用,那么会在requestHandler执行完成后自动回收
	// 如果是远程调用,那么在远程client将消息发送完成后自动回收)
	return clients[0].SendMessage(envelope)
}

// sendMessageWithFuture 发送消息(有回复)
func (ch *Handler) sendMessageWithFuture(nodeUID, serviceID, serviceMethod string, timeout time.Duration, isNode bool, reply interface{}, args ...interface{}) ([]interface{}, chaoserrors.CError) {
	serviceName, method, err := utils.SplitServiceMethod(serviceMethod)
	if err != nil {
		return nil, chaoserrors.NewErrCode(errcode.ServiceMethodError, err)
	}

	// 检查参数
	if reply != nil && reflect.TypeOf(reply).Kind() != reflect.Ptr {
		return nil, chaoserrors.NewErrCode(errcode.ServerError, "reply must be a pointer", nil)
	}

	// 获取客户端
	clients := endpoints.GetEndpointManager().GetServiceByID(nodeUID, serviceID, serviceName, false)
	//log.SysLogger.Debugf("sendMessageWithFuture nodeUID [%s] serviceName [%s] method [%s] clients name [%s]", nodeUID, serviceName, method, clients[0].GetName())
	if len(clients) == 0 {
		log.SysLogger.Errorf("can not find service nodeUID [%s] method [%s] id [%s]", nodeUID, serviceMethod, serviceID)
		return nil, chaoserrors.NewErrCode(errcode.ServerError, "can not find service", nil)
	} else if len(clients) > 1 {
		log.SysLogger.Errorf("can not call service nodeUID [%s] method [%s] id [%s] more than 1 service", nodeUID, serviceMethod, serviceID)
		return nil, chaoserrors.NewErrCode(errcode.ServerError, "call method more than 1 service", nil)
	}

	// 封装消息
	envelope := actor.NewMsgEnvelope()
	envelope.Method = method
	envelope.Sender = ch.GetPID()
	envelope.Receiver = clients[0].GetPID()
	envelope.ReqID = clients[0].GenerateSeq()
	envelope.IsRpc = isNode

	if isNode {
		if len(args) > 0 {
			envelope.Request = args[0]
		}
	} else {
		envelope.Request = args
	}

	// 向客户端发送消息
	future := clients[0].SendMessageWithFuture(envelope, timeout, nil)

	defer actor.ReleaseFuture(future)                 // 只要等到回复就可以释放
	defer clients[0].RemovePending(future.GetReqID()) // 不管有没有加入超时队列,都移除一次

	return future.Result()
}

func (ch *Handler) asyncSendMessage(nodeUID, serviceID, serviceMethod string, timeout time.Duration, completions []actor.CompletionFunc, isNode bool, args ...interface{}) (inf.CancelRpc, chaoserrors.CError) {
	if timeout == 0 {
		timeout = defaultTimeout
	}
	serviceName, method, err := utils.SplitServiceMethod(serviceMethod)
	if err != nil {
		return nil, chaoserrors.NewErrCode(errcode.ServiceMethodError, err)
	}

	// 获取客户端
	clients := endpoints.GetEndpointManager().GetServiceByID(nodeUID, serviceID, serviceName, false)
	if len(clients) == 0 {
		log.SysLogger.Errorf("can not find service method [%s] id [%s]", serviceMethod, serviceID)
		return endpoints.EmptyCancelRpc, chaoserrors.NewErrCode(errcode.ServerError, "can not find service", nil)
	} else if len(clients) > 1 {
		log.SysLogger.Errorf("can not call service method [%s] id [%s] more than 1 service", serviceMethod, serviceID)
		return endpoints.EmptyCancelRpc, chaoserrors.NewErrCode(errcode.ServerError, "call method more than 1 service", nil)
	}

	// 封装消息
	envelope := actor.NewMsgEnvelope()
	envelope.Method = method
	envelope.Sender = ch.GetPID()
	envelope.Receiver = clients[0].GetPID()
	envelope.ReqID = clients[0].GenerateSeq()
	envelope.AddCompletion(completions...)
	envelope.IsRpc = isNode

	if isNode {
		if len(args) > 0 {
			envelope.Request = args[0]
		}
	} else {
		envelope.Request = args
	}

	return clients[0].AsyncSendMessage(envelope, timeout, completions)
}

func (ch *Handler) castMessage(nodeUID, serviceID, serviceMethod string, args ...interface{}) chaoserrors.CError {
	serviceName, method, err := utils.SplitServiceMethod(serviceMethod)
	if err != nil {
		return chaoserrors.NewErrCode(errcode.ServiceMethodError, err)
	}
	// 获取客户端
	clients := endpoints.GetEndpointManager().GetServiceByID(nodeUID, serviceID, serviceName, true)
	if err != nil {
		return err
	}
	if len(clients) == 0 {
		log.SysLogger.Errorf("can not find service method [%s] id [%s]", serviceMethod, serviceID)
		return chaoserrors.NewErrCode(errcode.ServerError, "can not find service", nil)
	}

	// 发送到所有客户端
	for _, pClient := range clients {
		envelope := actor.NewMsgEnvelope()
		envelope.Method = method
		envelope.Receiver = pClient.GetPID()
		envelope.ReqID = pClient.GenerateSeq()
		envelope.SetArgs(args[0])
		if err := pClient.SendMessage(envelope); err != nil {
			log.SysLogger.Errorf("cast message[%s] error [%s]", method, err)
		}
	}
	return nil
}

func (ch *Handler) Call(serviceID, serviceMethod string, args ...interface{}) ([]interface{}, chaoserrors.CError) {
	nodeUID := endpoints.GetEndpointManager().GetUID()
	if serviceID == "" {
		serviceID = nodeUID
	}
	return ch.sendMessageWithFuture(nodeUID, serviceID, serviceMethod, defaultTimeout, false, nil, args...)
}

func (ch *Handler) CallNode(nodeUID, serviceID, serviceMethod string, in interface{}) ([]interface{}, chaoserrors.CError) {
	return ch.sendMessageWithFuture(nodeUID, serviceID, serviceMethod, defaultTimeout, true, in)
}

func (ch *Handler) CallWithTimeout(serviceID, serviceMethod string, timeout time.Duration, args ...interface{}) ([]interface{}, chaoserrors.CError) {
	nodeUID := endpoints.GetEndpointManager().GetUID()
	if serviceID == "" {
		serviceID = nodeUID
	}
	return ch.sendMessageWithFuture(nodeUID, serviceID, serviceMethod, timeout, false, nil, args...)
}

func (ch *Handler) CallNodeWithTimeout(nodeUID, serviceID, serviceMethod string, timeout time.Duration, in interface{}) ([]interface{}, chaoserrors.CError) {
	return ch.sendMessageWithFuture(nodeUID, serviceID, serviceMethod, timeout, true, in)
}

// AsyncCall 异步调用,会返回一个取消函数,调用取消函数可以取消回调调用 (回调函数的格式必须是func(resp []interface{}, errorlib.CError))
func (ch *Handler) AsyncCall(serviceID, serviceMethod string, timeout time.Duration, callbacks []actor.CompletionFunc, args ...interface{}) (inf.CancelRpc, chaoserrors.CError) {
	if len(callbacks) == 0 {
		return endpoints.EmptyCancelRpc, chaoserrors.NewErrCode(errcode.ServiceSyncCallbackFunError, "need callback func", nil)
	}
	nodeUID := endpoints.GetEndpointManager().GetUID()
	if serviceID == "" {
		serviceID = nodeUID
	}
	return ch.asyncSendMessage(nodeUID, serviceID, serviceMethod, timeout, callbacks, false, args...)
}

func (ch *Handler) AsyncCallNode(nodeUID, serviceID, serviceMethod string, timeout time.Duration, callbacks []actor.CompletionFunc, in interface{}) (inf.CancelRpc, chaoserrors.CError) {
	if len(callbacks) == 0 {
		return endpoints.EmptyCancelRpc, chaoserrors.NewErrCode(errcode.ServiceSyncCallbackFunError, "need callback func", nil)
	}
	return ch.asyncSendMessage(nodeUID, serviceID, serviceMethod, timeout, callbacks, true, in)
}

func (ch *Handler) Send(serviceID, serviceMethod string, args ...interface{}) chaoserrors.CError {
	nodeUID := endpoints.GetEndpointManager().GetUID()
	if serviceID == "" {
		serviceID = nodeUID
	}
	return ch.sendMessage(nodeUID, serviceID, serviceMethod, false, args...)
}

func (ch *Handler) SendNode(nodeUID, serviceID, serviceMethod string, in interface{}) chaoserrors.CError {
	return ch.sendMessage(nodeUID, serviceID, serviceMethod, true, in)
}

func (ch *Handler) CastNode(serviceMethod string, in interface{}) chaoserrors.CError {
	return ch.castMessage("", "", serviceMethod, in)
}

func (ch *Handler) HandleRequest(envelope *actor.MsgEnvelope) {
	defer func() {
		if err := recover(); err != nil {
			log.SysLogger.Errorf("service[%s] handle message panic: %v\ntrace:%s", ch.GetName(), err, debug.Stack())
			//errStr := fmt.Sprintf("service[%s] message method[%s] panic:%v", ch.GetName(), envelope.GetMethod(), errdef)
		}
	}()
	//log.SysLogger.Debugf("rpc handler -> receive request: %+v", envelope)
	var (
		paramList  []reflect.Value
		results    []reflect.Value
		resp       []interface{}
		cErr       chaoserrors.CError
		methodInfo *MethodInfo
		ok         bool
		params     []interface{}
	)
	if strings.HasPrefix(envelope.GetMethod(), "API") || strings.HasPrefix(envelope.GetMethod(), "API_") {
		methodInfo, ok = ch.mapFunctions[envelope.GetMethod()]
	} else if strings.HasPrefix(envelope.GetMethod(), "RPC") || strings.HasPrefix(envelope.GetMethod(), "RPC_") {
		methodInfo, ok = ch.mapRpcFunctions[envelope.GetMethod()]
	}

	if !ok {
		log.SysLogger.Errorf("service[%s] method[%s] not found\nenvelope:%+v", ch.GetName(), envelope.GetMethod(), envelope)
		cErr = chaoserrors.NewErrCode(errcode.ServiceMethodNotExist, fmt.Sprintf("method[%s] not found", envelope.GetMethod()), nil)
		log.SysLogger.Debugf("rpc handler -> method not found, errdef:%s", cErr)
		goto DoCallBack
	}

	// 准备参数
	//生成Call参数
	paramList = append(paramList, reflect.ValueOf(ch.GetRpcHandler())) //接受者

	switch envelope.Request.(type) {
	case []interface{}:
		params = envelope.Request.([]interface{})
	default:
		if envelope.Request != nil {
			params = append(params, envelope.Request)
		}
	}

	for _, param := range params {
		paramList = append(paramList, reflect.ValueOf(param))
	}

	//if envelope.RpcRequest != nil {
	//	paramList = append(paramList, reflect.ValueOf(envelope.RpcRequest))
	//}

	// 调用方法
	results = methodInfo.method.Func.Call(paramList)
	if len(results) != len(methodInfo.outs) {
		err := fmt.Errorf("method[%s] return value count not match", envelope.GetMethod())
		cErr = chaoserrors.NewErrCode(errcode.ServerError, err)
		//log.SysLogger.Debug("rpc handler -> call method error")
		goto DoCallBack
	}

	if methodInfo.outs == nil || 0 == len(methodInfo.outs) {
		// 没有返回值,直接返回
		//log.SysLogger.Debug("rpc handler -> empty response")
		goto DoCallBack
	}

	for i, t := range methodInfo.outs {
		result := results[i]
		if t.Kind() == reflect.Ptr ||
			t.Kind() == reflect.Interface ||
			t.Kind() == reflect.Func ||
			t.Kind() == reflect.Map ||
			t.Kind() == reflect.Slice ||
			t.Kind() == reflect.Chan {

			if t.Implements(reflect.TypeOf((*chaoserrors.CError)(nil)).Elem()) {
				if err, ok := result.Interface().(chaoserrors.CError); ok && err != nil {
					cErr = err
					goto DoCallBack
				}
			} else if t.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
				if err, ok := result.Interface().(error); ok && err != nil {
					cErr = chaoserrors.NewErrCode(errcode.ServerError, err)
					goto DoCallBack
				}
			} else {
				if result.IsNil() {
					resp = append(resp, nil)
				} else {
					resp = append(resp, result.Interface())
				}
			}
		} else {
			resp = append(resp, result.Interface())
		}
	}

	// 回复
DoCallBack:
	//log.SysLogger.Debug("rpc handler -> begin do response")
	if envelope.Sender != nil {
		respEnvelope := actor.NewMsgEnvelope()
		respEnvelope.SetReceiver(envelope.GetSender())
		respEnvelope.SetReply() // 是回复
		respEnvelope.SetReq(envelope.GetReq())
		respEnvelope.Method = envelope.Method
		if cErr != nil {
			respEnvelope.Err = cErr.Error()
		}

		if envelope.IsRpc {
			if len(resp) > 0 {
				respEnvelope.Request = resp[0]
			}
		} else {
			respEnvelope.Request = resp
		}

		respEnvelope.IsRpc = envelope.IsRpc

		client := endpoints.GetEndpointManager().GetClientByPID(envelope.Sender)
		if client == nil {
			// client已经不存在了,直接返回
			actor.ReleaseMsgEnvelope(respEnvelope)
			log.SysLogger.Errorf("service[%s] send message[%s] response to client failed, client not found: %+v", ch.GetName(), envelope.GetMethod(), envelope)
			return
		}

		if !envelope.IsRpc && client.FindPending(envelope.ReqID) == nil {
			// 已经超时了,不需要回复了
			actor.ReleaseMsgEnvelope(respEnvelope)
			log.SysLogger.Errorf("service[%s] send message[%s] response to client failed, request timeout: %+v", ch.GetName(), envelope.GetMethod(), envelope)
			return
		}

		if err := client.SendMessage(respEnvelope); err != nil {
			log.SysLogger.Errorf("service[%s] send message[%s] response to client failed, error: %v", ch.GetName(), envelope.GetMethod(), err)
		}
	} else {
		//log.SysLogger.Debug("rpc handler -> no need response")
	}
}

func (ch *Handler) HandlerResponse(envelope *actor.MsgEnvelope) {
	defer func() {
		if r := recover(); r != nil {
			log.SysLogger.Errorf("service[%s] handle message panic: %v\n trace:%s", ch.GetName(), r, debug.Stack())
		}
	}()
	log.SysLogger.Debugf("rpc response handler -> begin handle message: %+v", envelope)

	// 开始处理回复信息

	// 获取自己的client
	client := endpoints.GetEndpointManager().GetClientByPID(ch.GetPID())
	if client == nil {
		// 直接返回
		log.SysLogger.Errorf("service[%s] handle message[%s] response from client failed, client not found: %+v", ch.GetName(), envelope.GetMethod(), envelope)
		return
	}

	// rpc回调
	future := client.RemovePending(envelope.GetReq())
	if future != nil {
		if future.NeedCallback() {
			envelope.AddCompletion(future.GetCompletions()...)
		}
		future.SetResult(nil, nil)
	}

	envelope.RunCompletions()
}

func (ch *Handler) HandlerClientMsg(param interface{}) {
	defer func() {
		if err := recover(); err != nil {
			log.SysLogger.Errorf("service[%s] handle message panic: %v\ntrace:%s", ch.GetName(), err, debug.Stack())
		}
	}()

	// 客户端消息
	ch.rawFunction(param)
	return
}

func (ch *Handler) IsPrivate() bool {
	return len(ch.mapRpcFunctions) == 0
}
