package timer

import (
	"fmt"
	"github.com/njtc406/chaosengine/engine/inf"
	"github.com/njtc406/chaosengine/engine/utils/log"
	"github.com/njtc406/chaosengine/engine/utils/pool"
	"reflect"
	"runtime"
	"sync/atomic"
	"time"
)

type OnCloseTimer func(timer inf.ITimer)
type OnAddTimer func(timer inf.ITimer)

// Timer 定时器
type Timer struct {
	Id             uint64
	cancelled      int32           //是否关闭
	C              chan inf.ITimer //定时器管道
	interval       time.Duration   // 时间间隔（用于循环定时器）
	fireTime       time.Time       // 触发时间
	cb             func(uint64, interface{})
	cbEx           func(t inf.ITimer)
	cbCronEx       func(t inf.ITimer)
	cbTickerEx     func(t inf.ITimer)
	cbOnCloseTimer OnCloseTimer
	cronExpr       *CronExpr
	AdditionData   interface{} //定时器附加数据
	rOpen          bool        //是否重新打开

	ref bool
}

// Ticker 定时器
type Ticker struct {
	Timer
}

// Cron 定时器
type Cron struct {
	Timer
}

var timerPool = pool.NewPoolEx(make(chan pool.IPoolData, 10240), func() pool.IPoolData {
	return &Timer{}
})

var cronPool = pool.NewPoolEx(make(chan pool.IPoolData, 10240), func() pool.IPoolData {
	return &Cron{}
})

var tickerPool = pool.NewPoolEx(make(chan pool.IPoolData, 10240), func() pool.IPoolData {
	return &Ticker{}
})

func newTimer(d time.Duration, c chan inf.ITimer, cb func(uint64, interface{}), additionData interface{}) *Timer {
	timer := timerPool.Get().(*Timer)
	timer.AdditionData = additionData
	timer.C = c
	timer.fireTime = Now().Add(d)
	timer.cb = cb
	timer.interval = d
	timer.rOpen = false
	return timer
}

func releaseTimer(timer *Timer) {
	timerPool.Put(timer)
}

func newTicker() *Ticker {
	t := tickerPool.Get().(*Ticker)
	return t
}

func releaseTicker(ticker *Ticker) {
	tickerPool.Put(ticker)
}

func newCron() *Cron {
	cron := cronPool.Get().(*Cron)
	return cron
}

func releaseCron(cron *Cron) {
	cronPool.Put(cron)
}

// Dispatcher one dispatcher per goroutine (goroutine not safe)
type Dispatcher struct {
	ChanTimer chan inf.ITimer
}

func (t *Timer) GetId() uint64 {
	return t.Id
}

func (t *Timer) GetFireTime() time.Time {
	return t.fireTime
}

func (t *Timer) Open(bOpen bool) {
	t.rOpen = bOpen
}

func (t *Timer) AppendChannel(timer inf.ITimer) {
	t.C <- timer
}

func (t *Timer) IsOpen() bool {
	return t.rOpen
}

func (t *Timer) Do() {
	defer func() {
		if r := recover(); r != nil {
			// 纪录日志
			log.SysLogger.Error(r)
		}
	}()

	if t.IsActive() == false {
		if t.cbOnCloseTimer != nil {
			t.cbOnCloseTimer(t)
		}

		releaseTimer(t)
		return
	}

	if t.cb != nil {
		t.cb(t.Id, t.AdditionData)
	} else if t.cbEx != nil {
		t.cbEx(t)
	}

	if t.rOpen == false {
		if t.cbOnCloseTimer != nil {
			t.cbOnCloseTimer(t)
		}
		releaseTimer(t)
	}
}

func (t *Timer) SetupTimer(now time.Time) error {
	t.fireTime = now.Add(t.interval)
	if SetupTimer(t) == nil {
		return fmt.Errorf("failed to install timer")
	}
	return nil
}

func (t *Timer) GetInterval() time.Duration {
	return t.interval
}

func (t *Timer) Cancel() {
	atomic.StoreInt32(&t.cancelled, 1)
}

// 判断定时器是否已经取消
func (t *Timer) IsActive() bool {
	return atomic.LoadInt32(&t.cancelled) == 0
}

func (t *Timer) GetName() string {
	if t.cb != nil {
		return runtime.FuncForPC(reflect.ValueOf(t.cb).Pointer()).Name()
	} else if t.cbEx != nil {
		return runtime.FuncForPC(reflect.ValueOf(t.cbEx).Pointer()).Name()
	}

	return ""
}

var emptyTimer Timer

func (t *Timer) Reset() {
	*t = emptyTimer
}

func (t *Timer) IsRef() bool {
	return t.ref
}

func (t *Timer) Ref() {
	t.ref = true
}

func (t *Timer) UnRef() {
	t.ref = false
}

func (c *Cron) Reset() {
	c.Timer.Reset()
}

func (c *Cron) Do() {
	defer func() {
		if r := recover(); r != nil {
			log.SysLogger.Error(r)
		}
	}()

	if c.IsActive() == false {
		if c.cbOnCloseTimer != nil {
			c.cbOnCloseTimer(c)
		}
		releaseCron(c)
		return
	}

	now := Now()
	nextTime := c.cronExpr.Next(now)
	if nextTime.IsZero() {
		c.cbCronEx(c)
		return
	}

	if c.cb != nil {
		c.cb(c.Id, c.AdditionData)
	} else if c.cbCronEx != nil {
		c.cbCronEx(c)
	}

	if c.IsActive() == true {
		c.interval = nextTime.Sub(now)
		c.fireTime = now.Add(c.interval)
		SetupTimer(c)
	} else {
		if c.cbOnCloseTimer != nil {
			c.cbOnCloseTimer(c)
		}
		releaseCron(c)
		return
	}
}

func (c *Cron) IsRef() bool {
	return c.ref
}

func (c *Cron) Ref() {
	c.ref = true
}

func (c *Cron) UnRef() {
	c.ref = false
}

func (c *Ticker) Do() {
	defer func() {
		if r := recover(); r != nil {
			log.SysLogger.Error(r)
		}
	}()

	if c.IsActive() == false {
		if c.cbOnCloseTimer != nil {
			c.cbOnCloseTimer(c)
		}

		releaseTicker(c)
		return
	}

	if c.cb != nil {
		c.cb(c.Id, c.AdditionData)
	} else if c.cbTickerEx != nil {
		c.cbTickerEx(c)
	}

	if c.IsActive() == true {
		c.fireTime = Now().Add(c.interval)
		SetupTimer(c)
	} else {
		if c.cbOnCloseTimer != nil {
			c.cbOnCloseTimer(c)
		}
		releaseTicker(c)
	}
}

func (c *Ticker) Reset() {
	c.Timer.Reset()
}

func (c *Ticker) IsRef() bool {
	return c.ref
}

func (c *Ticker) Ref() {
	c.ref = true
}

func (c *Ticker) UnRef() {
	c.ref = false
}

func NewDispatcher(l int) *Dispatcher {
	dispatcher := new(Dispatcher)
	dispatcher.ChanTimer = make(chan inf.ITimer, l)
	return dispatcher
}

func (dp *Dispatcher) AfterFunc(d time.Duration, cb func(uint64, interface{}), cbEx func(inf.ITimer), onTimerClose OnCloseTimer, onAddTimer OnAddTimer) *Timer {
	timer := newTimer(d, dp.ChanTimer, nil, nil)
	timer.cb = cb
	timer.cbEx = cbEx
	timer.cbOnCloseTimer = onTimerClose

	t := SetupTimer(timer)
	if onAddTimer != nil && t != nil {
		onAddTimer(t)
	}

	return timer
}

func (dp *Dispatcher) CronFunc(cronExpr *CronExpr, cb func(uint64, interface{}), cbEx func(inf.ITimer), onTimerClose OnCloseTimer, onAddTimer OnAddTimer) *Cron {
	now := Now()
	nextTime := cronExpr.Next(now)
	if nextTime.IsZero() {
		return nil
	}

	cron := newCron()
	cron.cb = cb
	cron.cbCronEx = cbEx
	cron.cbOnCloseTimer = onTimerClose
	cron.cronExpr = cronExpr
	cron.C = dp.ChanTimer
	cron.interval = nextTime.Sub(now)
	cron.fireTime = nextTime
	SetupTimer(cron)
	onAddTimer(cron)
	return cron
}

func (dp *Dispatcher) TickerFunc(d time.Duration, cb func(uint64, interface{}), cbEx func(inf.ITimer), onTimerClose OnCloseTimer, onAddTimer OnAddTimer) *Ticker {
	t := newTicker()
	t.C = dp.ChanTimer
	t.fireTime = Now().Add(d)
	t.cb = cb
	t.cbTickerEx = cbEx
	t.interval = d

	// callback
	SetupTimer(t)
	onAddTimer(t)

	return t
}
