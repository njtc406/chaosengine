// Package discovery
// @Title  服务发现
// @Description  服务发现
// @Author  yr  2024/8/29 下午3:42
// @Update  yr  2024/8/29 下午3:42
package discovery

import (
	"context"
	"github.com/njtc406/chaosengine/engine/cluster/config"
	"github.com/njtc406/chaosengine/engine/dto"
	"github.com/njtc406/chaosengine/engine/errdef"
	"github.com/njtc406/chaosengine/engine/inf"
	nodeConfig "github.com/njtc406/chaosengine/engine/node/config"
	"path"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/njtc406/chaosengine/engine/actor"
	"github.com/njtc406/chaosengine/engine/event"
	"github.com/njtc406/chaosengine/engine/utils/asynclib"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/pool"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

const minWatchTTL = time.Second * 3

var locker sync.Mutex
var discoveryMap = map[string]inf.IDiscovery{
	"etcd": &EtcdDiscovery{},
}

func Register(name string, discovery inf.IDiscovery) {
	locker.Lock()
	discoveryMap[name] = discovery
	locker.Unlock()
}

func CreateDiscovery(name string) inf.IDiscovery {
	locker.Lock()
	defer locker.Unlock()
	return discoveryMap[name]
}

type EtcdDiscovery struct {
	inf.IProcessor
	inf.IHandler

	closed     bool
	watching   int32
	conf       *config.EtcdConf
	client     *clientv3.Client
	mapWatcher *sync.Map
	t          *time.Timer
}

func NewDiscovery() *EtcdDiscovery {
	return &EtcdDiscovery{}
}

func (d *EtcdDiscovery) Init(conf interface{}, eventProcessor inf.IProcessor) (err error) {
	if conf == nil {
		return errdef.DiscoveryConfNotFound
	}
	d.IProcessor = eventProcessor
	d.IHandler = event.NewHandler()
	d.IHandler.Init(d.IProcessor)
	d.conf = conf.(*config.EtcdConf)
	d.mapWatcher = &sync.Map{}

	if d.conf.TTL < minWatchTTL {
		d.conf.TTL = minWatchTTL
	}

	return d.conn()
}

func (d *EtcdDiscovery) conn() (err error) {
	if len(nodeConfig.Conf.ETCDConf.EtcdEndPoints) == 0 {
		log.SysLogger.Info("etcd end points is empty")
		return
	}

	d.client, err = clientv3.New(clientv3.Config{
		Endpoints:   nodeConfig.Conf.ETCDConf.EtcdEndPoints,
		DialTimeout: nodeConfig.Conf.ETCDConf.DialTimeout,
		Username:    nodeConfig.Conf.ETCDConf.UserName,
		Password:    nodeConfig.Conf.ETCDConf.Password,
		Logger:      zap.NewNop(),
	})
	if err == nil {
		log.SysLogger.Info("etcd connect success")
	}

	return
}

func (d *EtcdDiscovery) Start() {
	if d.client == nil {
		return
	}
	if !atomic.CompareAndSwapInt32(&d.watching, 0, 1) {
		return
	}
	d.IProcessor.RegEventReceiverFunc(event.SysEventServiceReg, d.IHandler, d.addService)
	asynclib.Go(d.watcher)
	tp := time.AfterFunc(time.Second, d.getAll)
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&d.t)), unsafe.Pointer(&tp))
}

func (d *EtcdDiscovery) Close() {
	if d.client != nil {
		// 通知所有goroutine和定时器关闭
		d.closed = true
		d.closeWatchers()    // 所有本地服务下线
		_ = d.client.Close() // 关闭连接
		d.client = nil
	}
	atomic.StoreInt32(&d.watching, 0)
}

func (d *EtcdDiscovery) closeWatchers() {
	d.mapWatcher.Range(func(key, value interface{}) bool {
		if watcher, ok := value.(*watcherInfo); ok {
			// 删除租约
			leaseID := watcher.getLeaseID()
			if leaseID <= 0 {
				// 这里可能是由于服务下线,导致已经删除了
				watcher.Close()             // 关闭watcher
				d.mapWatcher.Delete(key)    // 删除本地缓存
				releaseWatcherInfo(watcher) // 回收
				return true
			}
			_, err := d.client.Revoke(context.Background(), watcher.getLeaseID())
			if err != nil {
				log.SysLogger.Errorf("etcd revoke lease error: %v", err)
			}
			watcher.Close()             // 关闭watcher
			d.mapWatcher.Delete(key)    // 删除本地缓存
			releaseWatcherInfo(watcher) // 回收
		}

		return true
	})
}

func (d *EtcdDiscovery) getAll() {
	if d.client == nil {
		return
	}
	// 获取当前所有服务
	resp, err := d.client.Get(context.Background(), d.conf.Path, clientv3.WithPrefix())
	if err != nil {
		log.SysLogger.Errorf("etcd get service error: %v", err)
	} else if len(resp.Kvs) > 0 {
		for _, kv := range resp.Kvs {
			// 注册或者修改服务
			ent := event.NewEvent()
			ent.Type = event.SysEventETCDPut
			ent.Data = kv
			if err = d.PushEvent(ent); err != nil {
				log.SysLogger.Errorf("etcd put service error: %v", err)
			}
		}
	}
}

func (d *EtcdDiscovery) watcher() {
	log.SysLogger.Info("etcd watcher start")
	d.watch()
	log.SysLogger.Info("etcd watcher stop")
}

func (d *EtcdDiscovery) watch() {
	log.SysLogger.Info("etcd watcher running")
	defer func() {
		if err := recover(); err != nil {
			log.SysLogger.Errorf("etcd watch error: %v", err)
		}
	}()

	changes := d.client.Watch(context.Background(), d.conf.Path, clientv3.WithPrefix())

	for !d.closed {
		select {
		case resp := <-changes:
			for _, ev := range resp.Events {
				var ent *event.Event
				switch ev.Type {
				case clientv3.EventTypePut:
					// 注册或者修改服务
					ent = event.NewEvent()
					ent.Type = event.SysEventETCDPut
					ent.Data = ev.Kv
				case clientv3.EventTypeDelete:
					// 注销服务
					ent = event.NewEvent()
					ent.Type = event.SysEventETCDDel
					ent.Data = ev.Kv
				default:
					continue
				}
				d.PushEvent(ent)
			}
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}

	return
}

func (d *EtcdDiscovery) startKeepalive(watcher *watcherInfo) {
	asynclib.Go(func() {
		for !watcher.closed {
			d.keepaliveForever(watcher)
		}
	})
}

func (d *EtcdDiscovery) keepaliveForever(watcher *watcherInfo) {
	log.SysLogger.Info("watcher keepalive start")
	defer func() {
		if err := recover(); err != nil {
			log.SysLogger.Errorf("etcd keepalive error: %v", err)
		}
	}()
	leaseID := watcher.getLeaseID()
	if leaseID <= 0 {
		// 这里可能是由于服务下线,导致已经删除了
		return
	}
	kaRespCh, err := d.client.KeepAlive(context.Background(), leaseID)
	if err != nil {
		log.SysLogger.Errorf("etcd keepalive error: %v", err)
		return
	}

	for !watcher.closed {
		select {
		case kaResp, ok := <-kaRespCh:
			if !ok || kaResp == nil {
				log.SysLogger.Errorf("etcd keepalive error: %v", kaResp)
				// 尝试重连
				return
			}
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func (d *EtcdDiscovery) addService(ev inf.IEvent) {
	if d.closed || d.client == nil {
		return
	}
	if ent, ok := ev.(*event.Event); ok {
		if ent.Type == event.SysEventServiceReg {
			pid := ent.Data.(*actor.PID)
			watcher := d.getWatcherInfo(pid)
			if watcher == nil {
				watcher = newWatcherInfo()
				watcher.pid = pid
				leaseID, err := d.newLeaseID()
				if err != nil {
					log.SysLogger.Errorf("etcd create lease error: %v", err)
					return
				}
				watcher.setLeaseID(leaseID)
				d.addWatcherInfo(watcher)
			}
			// 注册服务
			fullPath := path.Join(d.conf.Path, pid.GetServiceUid())
			log.SysLogger.Debugf("etcd lease id: %d", watcher.getLeaseID())
			log.SysLogger.Debugf("etcd full path: %s", fullPath)
			pidByte, err := protojson.Marshal(pid)
			if err != nil {
				log.SysLogger.Errorf("etcd marshal pid error: %v", err)
			}
			_, err = d.client.Put(context.Background(), fullPath, string(pidByte), clientv3.WithLease(watcher.getLeaseID()))
			if err != nil {
				log.SysLogger.Errorf("etcd register service error: %v", err)
			}

			log.SysLogger.Infof("etcd register service success: %v", pid.String())
			asynclib.Go(func() {
				d.keepaliveForever(watcher)
			}) // 启动心跳
		}
	}
}
func (d *EtcdDiscovery) removeService(ev inf.IEvent) {
	if d.closed {
		return
	}
	if ent, ok := ev.(*event.Event); ok {
		if ent.Type == event.SysEventServiceDis {
			pid := ent.Data.(*actor.PID)
			watcher := d.getWatcherInfo(pid)
			if watcher != nil {
				// 注销服务
				_, err := d.client.Delete(context.Background(), path.Join(d.conf.Path, pid.GetServiceUid()))
				if err != nil {
					log.SysLogger.Errorf("etcd unregister service error: %v", err)
				}

				watcher.Close()
				d.mapWatcher.Delete(pid.GetServiceUid())
				d.removeWatcherInfo(pid.GetServiceUid())
			}
		}
	}
}

func (d *EtcdDiscovery) newLeaseID() (clientv3.LeaseID, error) {
	resp, err := d.client.Grant(context.Background(), int64(d.conf.TTL))
	if err != nil {
		return 0, err
	}
	return resp.ID, nil
}

func (d *EtcdDiscovery) getWatcherInfo(pid *actor.PID) *watcherInfo {
	if watcher, ok := d.mapWatcher.Load(pid.GetServiceUid()); ok {
		return watcher.(*watcherInfo)
	}
	return nil
}

func (d *EtcdDiscovery) addWatcherInfo(watcher *watcherInfo) {
	d.mapWatcher.Store(watcher.pid.GetServiceUid(), watcher)
}
func (d *EtcdDiscovery) removeWatcherInfo(uid string) {
	if watcher, ok := d.mapWatcher.LoadAndDelete(uid); ok {
		// 移除租约
		_, err := d.client.Revoke(context.Background(), watcher.(*watcherInfo).getLeaseID())
		if err != nil {
			log.SysLogger.Errorf("etcd revoke lease error: %v", err)
		}

		releaseWatcherInfo(watcher.(*watcherInfo))
	}
}

type watcherInfo struct {
	dto.DataRef
	pid     *actor.PID
	leaseID clientv3.LeaseID
	closed  bool
}

func (w *watcherInfo) Reset() {
	w.pid = nil
	w.setLeaseID(0)
	w.closed = false
}

func (w *watcherInfo) getLeaseID() clientv3.LeaseID {
	return (clientv3.LeaseID)(atomic.LoadInt64((*int64)(&w.leaseID)))
}

func (w *watcherInfo) setLeaseID(leaseID clientv3.LeaseID) {
	atomic.StoreInt64((*int64)(&w.leaseID), (int64)(leaseID))
}

func (w *watcherInfo) Close() {
	w.closed = true
}

var watcherPool = pool.NewPoolEx(make(chan pool.IPoolData, 1024), func() pool.IPoolData {
	return &watcherInfo{}
})

func newWatcherInfo() *watcherInfo {
	return watcherPool.Get().(*watcherInfo)
}

func releaseWatcherInfo(watcher *watcherInfo) {
	if watcher != nil {
		watcherPool.Put(watcher)
	}
}
