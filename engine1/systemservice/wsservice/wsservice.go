// Package wsservice
// @Title
// @Description  wsservice
// @Author sly 2024/9/10
// @Created sly 2024/9/10
package wsservice

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/njtc406/chaosengine/engine1/errdef/errcode"
	"github.com/njtc406/chaosengine/engine1/event"
	"github.com/njtc406/chaosengine/engine1/log"
	"github.com/njtc406/chaosengine/engine1/network"
	"github.com/njtc406/chaosengine/engine1/network/processor"
	"github.com/njtc406/chaosengine/engine1/service"
	"github.com/njtc406/chaosutil/chaoserrors"
	"strings"
	"sync"
)

type WSService struct {
	service.Service
	wsServer network.WSServer

	mapClientLocker sync.RWMutex
	mapClient       map[string]*WSClient
	process         processor.IProcessor
}

type WSPackType int8

const (
	WPT_Connected    WSPackType = 0
	WPT_DisConnected WSPackType = 1
	WPT_Pack         WSPackType = 2
	WPT_UnknownPack  WSPackType = 3
)

const Default_WS_MaxConnNum = 3000
const Default_WS_PendingWriteNum = 10000
const Default_WS_MaxMsgLen = 65535

type WSClient struct {
	id        string
	wsConn    *network.WSConn
	wsService *WSService
}

type WSPack struct {
	Type         WSPackType //0表示连接 1表示断开 2表示数据
	MsgProcessor processor.IProcessor
	ClientId     string
	Data         interface{}
}

func (ws *WSService) getConf() *network.WSServer {
	return ws.GetServiceCfg().(*network.WSServer)
}

func (ws *WSService) OnInit() chaoserrors.CError {

	iConfig := ws.getConf()
	if iConfig == nil {
		log.SysLogger.Errorf("配置文件不存在")
		return chaoserrors.NewErrCode(errcode.ConfNotExist, "配置文件不存在", nil)
	}

	ws.wsServer.Addr = iConfig.Addr
	ws.wsServer.MaxConnNum = iConfig.MaxConnNum
	ws.wsServer.PendingWriteNum = iConfig.PendingWriteNum
	ws.wsServer.MaxMsgLen = iConfig.MaxMsgLen
	ws.wsServer.MaxConnNum = iConfig.MaxConnNum

	ws.mapClient = make(map[string]*WSClient, ws.wsServer.MaxConnNum)
	ws.wsServer.NewAgent = ws.NewWSClient
	ws.wsServer.Start()
	return nil
}

func (ws *WSService) SetMessageType(messageType int) {
	ws.wsServer.SetMessageType(messageType)
}

func (ws *WSService) WSEventHandler(ev event.IEvent) {
	pack := ev.(*event.Event).Data.(*WSPack)
	switch pack.Type {
	case WPT_Connected:
		pack.MsgProcessor.ConnectedRoute(pack.ClientId)
	case WPT_DisConnected:
		pack.MsgProcessor.DisConnectedRoute(pack.ClientId)
	case WPT_UnknownPack:
		pack.MsgProcessor.UnknownMsgRoute(pack.ClientId, pack.Data)
	case WPT_Pack:
		pack.MsgProcessor.MsgRoute(pack.ClientId, pack.Data)
	}
}

func (ws *WSService) SetProcessor(process processor.IProcessor, handler event.IHandler) {
	ws.process = process
	ws.RegEventReceiverFunc(event.SysEventWebsocket, handler, ws.WSEventHandler)
}

func (ws *WSService) NewWSClient(conn *network.WSConn) network.Agent {
	ws.mapClientLocker.Lock()
	defer ws.mapClientLocker.Unlock()

	uuId, _ := uuid.NewUUID()
	clientId := strings.ReplaceAll(uuId.String(), "-", "")
	pClient := &WSClient{wsConn: conn, id: clientId}
	pClient.wsService = ws
	ws.mapClient[clientId] = pClient
	return pClient
}

func (slf *WSClient) GetId() string {
	return slf.id
}

func (slf *WSClient) Run() {
	slf.wsService.NotifyEvent(&event.Event{Type: event.SysEventWebsocket, Data: &WSPack{ClientId: slf.id, Type: WPT_Connected, MsgProcessor: slf.wsService.process}})
	for {
		bytes, err := slf.wsConn.ReadMsg()
		if err != nil {
			log.SysLogger.Debug("read client id %s is error:%+v", slf.id, err)
			break
		}
		data, err := slf.wsService.process.Unmarshal(slf.id, bytes)
		if err != nil {
			slf.wsService.NotifyEvent(&event.Event{Type: event.SysEventWebsocket, Data: &WSPack{ClientId: slf.id, Type: WPT_UnknownPack, Data: bytes, MsgProcessor: slf.wsService.process}})
			continue
		}
		slf.wsService.NotifyEvent(&event.Event{Type: event.SysEventWebsocket, Data: &WSPack{ClientId: slf.id, Type: WPT_Pack, Data: data, MsgProcessor: slf.wsService.process}})
	}
}

func (slf *WSClient) OnClose() {
	slf.wsService.NotifyEvent(&event.Event{Type: event.SysEventWebsocket, Data: &WSPack{ClientId: slf.id, Type: WPT_DisConnected, MsgProcessor: slf.wsService.process}})
	slf.wsService.mapClientLocker.Lock()
	defer slf.wsService.mapClientLocker.Unlock()
	delete(slf.wsService.mapClient, slf.GetId())
}

func (ws *WSService) SendMsg(clientid string, msg interface{}) error {
	ws.mapClientLocker.Lock()
	client, ok := ws.mapClient[clientid]
	if ok == false {
		ws.mapClientLocker.Unlock()
		return fmt.Errorf("client %s is disconnect!", clientid)
	}

	ws.mapClientLocker.Unlock()
	bytes, err := ws.process.Marshal(clientid, msg)
	if err != nil {
		return err
	}
	return client.wsConn.WriteMsg(bytes)
}

func (ws *WSService) Close(clientid string) {
	ws.mapClientLocker.Lock()
	defer ws.mapClientLocker.Unlock()

	client, ok := ws.mapClient[clientid]
	if ok == false {
		return
	}

	if client.wsConn != nil {
		client.wsConn.Close()
	}

	return
}

func (ws *WSService) recyclerReaderBytes(data []byte) {
}
