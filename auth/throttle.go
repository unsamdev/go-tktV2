package auth

import (
	"github.com/fvk113/go-tkt-convenios/util"
	"sync"
	"time"
)

type ThrottleConfig struct {
	FailCount          *int `json:"failCount"`
	EvaluationInterval *int `json:"evaluationInterval"`
	DenialInterval     *int `json:"denialInterval"`
}

type ThrottleEntry struct {
	Key        string
	StartTime  time.Time
	FailCount  int
	Denied     bool
	DenialTime *time.Time
}

type ThrottleManager struct {
	throttleConfig   ThrottleConfig
	entryMap         map[string]*ThrottleEntry
	mux              *sync.Mutex
	evaluationWindow time.Duration
	denialDuration   time.Duration
}

func (o *ThrottleConfig) Validate() {
	if o.FailCount == nil {
		panic("Invalid failCount")
	}
	if o.EvaluationInterval == nil {
		panic("Invalid evaluationInterval")
	}
	if o.DenialInterval == nil {
		panic("Invalid denialInterval")
	}
}

func (o *ThrottleManager) registerFail(key string) bool {
	o.mux.Lock()
	defer o.mux.Unlock()
	te := o.entryMap[key]
	if te == nil {
		te = &ThrottleEntry{Key: key, StartTime: time.Now(), FailCount: 0}
		o.entryMap[key] = te
	}
	te.FailCount++
	t1 := time.Now()
	d := t1.Sub(te.StartTime)
	inRange := d < o.evaluationWindow
	te.Denied = te.Denied || (inRange && d <= o.evaluationWindow)
	if te.Denied {
		te.DenialTime = util.PTime(time.Now())
	}
	if !inRange {
		te.FailCount = 1
		te.StartTime = time.Now()
	}
	return te.Denied
}

func (o *ThrottleManager) isDenied(key string) bool {
	o.mux.Lock()
	o.mux.Unlock()
	te := o.entryMap[key]
	if te == nil {
		return false
	} else {
		o.cleanup(te)
		return te.Denied
	}
}

func (o *ThrottleManager) cleanup(te *ThrottleEntry) {
	if te.Denied {
		te.Denied = !te.DenialTime.Before(te.DenialTime.Add(o.denialDuration))
		te.FailCount = 0
	}
	if te.FailCount == 0 {
		delete(o.entryMap, te.Key)
	}
}

func NewThrottleManager(throttleConfig ThrottleConfig) *ThrottleManager {
	w := time.Duration(*throttleConfig.EvaluationInterval) * time.Minute
	d := time.Duration(*throttleConfig.DenialInterval) * time.Minute
	return &ThrottleManager{throttleConfig: throttleConfig, entryMap: make(map[string]*ThrottleEntry), mux: &sync.Mutex{},
		evaluationWindow: w, denialDuration: d}
}
