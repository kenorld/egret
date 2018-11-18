package jobs

import (
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/kenorld/egret"
	"github.com/kenorld/egret/core/logging"
	"github.com/kenorld/egret/cron"
	"go.uber.org/zap"
)

type Job struct {
	Name    string
	inner   cron.Job
	status  uint32
	running sync.Mutex
}

const UNNAMED = "(unnamed)"

func New(job cron.Job) *Job {
	name := reflect.TypeOf(job).Name()
	if name == "Func" {
		name = UNNAMED
	}
	return &Job{
		Name:  name,
		inner: job,
	}
}

func (j *Job) Status() string {
	if atomic.LoadUint32(&j.status) > 0 {
		return "RUNNING"
	}
	return "IDLE"
}

func (j *Job) Run() {
	// If the job panics, just print a stack trace.
	// Don't let the whole process die.
	defer func() {
		if err := recover(); err != nil {
			if egretError := egret.NewErrorFromPanic(err); egretError != nil {
				logging.Logger.Error("error in job", zap.Any("error", err), zap.String("stack", egretError.Stack))
			} else {
				logging.Logger.Error("error in job", zap.Any("error", err), zap.String("stack", string(debug.Stack())))
			}
		}
	}()

	if !selfConcurrent {
		j.running.Lock()
		defer j.running.Unlock()
	}

	if workPermits != nil {
		workPermits <- struct{}{}
		defer func() { <-workPermits }()
	}

	atomic.StoreUint32(&j.status, 1)
	defer atomic.StoreUint32(&j.status, 0)

	j.inner.Run()
}
