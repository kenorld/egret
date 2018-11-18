package jobs

import (
	"github.com/kenorld/egret"
	"github.com/kenorld/egret/cron"
)

const DefaultJobPoolSize = 10

var (
	// Singleton instance of the underlying job scheduler.
	MainCron *cron.Cron

	// This limits the number of jobs allowed to run concurrently.
	workPermits chan struct{}

	// Is a single job allowed to run concurrently with itself?
	selfConcurrent bool
)

func init() {
	MainCron = cron.New()
	egret.OnAppStart(func() {
		if size := egret.Config.GetIntDefault("jobs.pool", DefaultJobPoolSize); size > 0 {
			workPermits = make(chan struct{}, size)
		}
		selfConcurrent = egret.Config.GetBoolDefault("jobs.self_concurrent", false)
		MainCron.Start()
	})
}
