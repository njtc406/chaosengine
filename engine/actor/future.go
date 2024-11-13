// Package actor
// @Title  数据结果
// @Description  用于同步调用时的数据返回
// @Author  yr  2024/9/2 下午3:40
// @Update  yr  2024/9/2 下午3:40
package actor

// TODO 这个之后做rpc等待队列的时候再看怎么弄，现在这个稍微有点不明不白的

//type Future struct {
//	synclib.DataRef
//
//	sender      *PID
//	done        chan struct{}
//	result      interface{}
//	err         error
//	completions []CompletionFunc
//	//t           *time.Timer
//	timeout time.Duration
//	method  string
//	reqID   uint64
//}
//
//func (f *Future) Reset() {
//	f.sender = nil
//	if f.done != nil && len(f.done) > 0 {
//		<-f.done
//	}
//	if f.done == nil {
//		f.done = make(chan struct{}, 1)
//	}
//	f.result = nil
//	f.err = nil
//	f.completions = nil
//	f.reqID = 0
//	f.timeout = 0
//	f.method = ""
//}
//
//func (f *Future) GetSender() *PID {
//	return f.sender
//}
//
//func (f *Future) wait() {
//	<-f.done
//}
//
//func (f *Future) Wait() {
//	f.wait()
//}
//
//func (f *Future) Result() (interface{}, error) {
//	f.wait()
//	return f.result, f.err
//}
//
//func (f *Future) SetResult(res interface{}, err error) {
//	f.result = res
//	f.err = err
//	f.done <- struct{}{}
//}
//
//func (f *Future) Done() {
//	f.done <- struct{}{}
//}
//
//func (f *Future) SetSender(sender *PID) {
//	f.sender = sender
//}
//
//func (f *Future) SetCompletions(cbs ...CompletionFunc) {
//	f.completions = append(f.completions, cbs...)
//}
//
//func (f *Future) GetCompletions() []CompletionFunc {
//	return f.completions
//}
//
//func (f *Future) NeedCallback() bool {
//	return len(f.completions) > 0
//}
//
//func (f *Future) SetTimeout(timeout time.Duration) {
//	if timeout > 0 {
//		f.timeout = timeout
//	}
//}
//
//func (f *Future) GetTimeout() time.Duration {
//	return f.timeout
//}
//
//func (f *Future) SetMethod(method string) {
//	f.method = method
//}
//
//func (f *Future) GetMethod() string {
//	return f.method
//}
//
//func (f *Future) SetReqID(reqID uint64) {
//	f.reqID = reqID
//}
//func (f *Future) GetReqID() uint64 {
//	return f.reqID
//}
//
//var pool = synclib.NewPoolEx(make(chan synclib.IPoolData, 10240), func() synclib.IPoolData {
//	return &Future{done: make(chan struct{}, 1)}
//})
//
//var futureCount int64
//
//func NewFuture() *Future {
//	//log.SysLogger.Debugf("===============future count add: %d", atomic.AddInt64(&futureCount, 1))
//	return pool.Get().(*Future)
//}
//
//func ReleaseFuture(f *Future) {
//	if f != nil {
//		//log.SysLogger.Debugf("===============future count sub: %d", atomic.AddInt64(&futureCount, -1))
//		pool.Put(f)
//	}
//}
