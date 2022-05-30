package debug

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type errMap map[string]uint64

type Reporter struct {
	sync.RWMutex
	debugger      Debugging
	id            uint32
	currLatency   uint64
	currDone      uint64
	currAttempts  uint64
	currFail      uint64
	totalDone     uint64
	totalAttempts uint64
	totalFails    uint64
	errors        errMap
	accErrors     errMap
}

func NewReporter(debugger Debugging) *Reporter {
	return &Reporter{
		debugger:  debugger,
		errors:    make(map[string]uint64),
		accErrors: make(map[string]uint64),
	}
}

func (r *Reporter) AddStat(errs []error, attempts int) {
	atomic.AddUint64(&r.currAttempts, uint64(attempts))

	if len(errs) == 0 {
		return
	}

	r.Lock()
	defer r.Unlock()
	for _, err := range errs {
		r.errors[ParseError(err)]++
	}
}
func (r *Reporter) Done() {
	atomic.AddUint64(&r.currDone, 1)
}

func (r *Reporter) Fail() {
	atomic.AddUint64(&r.currFail, 1)
}

func (r *Reporter) Report(ctx context.Context, interval time.Duration) {
	go func() {
		seconds := int(interval.Seconds())
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ctx = NewDebugCtx(ctx, fmt.Sprintf("%d-Report", atomic.AddUint32(&r.id, 1)))
				done, attempts, err := atomic.SwapUint64(&r.currDone, 0), atomic.SwapUint64(&r.currAttempts, 0), atomic.SwapUint64(&r.currFail, 0)
				r.totalDone, r.totalAttempts, r.totalFails = r.totalDone+done, r.totalAttempts+attempts, r.totalFails+err

				r.Lock()
				errors := r.errors
				r.errors = make(map[string]uint64)
				r.Unlock()

				for msg, count := range errors {
					r.accErrors[msg] += count
				}
				r.debugger.Debugf(ctx, "Last %d seconds - Response: %7d, Average Attempts: %.2f, Error Rate: %.3f, ErrorMap: %s",
					seconds, done, float64(attempts)/float64(done), float64(err)/float64(done), errors.summary())
				r.debugger.Debugf(ctx, "     Total      - Response: %7d, Average Attempts: %.2f, Error Rate: %.3f, ErrorMap: %s",
					r.totalDone, float64(r.totalAttempts)/float64(r.totalDone), float64(r.totalFails)/float64(r.totalDone), r.accErrors.summary())
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()
}

func (r *errMap) summary() string {
	if r == nil || len(*r) == 0 {
		return "No Errors"
	}

	summary := []string{""}
	for msg, count := range *r {
		summary = append(summary, fmt.Sprintf("%s%d", r.padding(msg), count))
	}
	return Padding(strings.Join(summary, "\n"), 8)
}

func (r *errMap) padding(msg string) string {
	pad := 70 - len(msg)
	if pad < 0 {
		pad = 0
	}

	return fmt.Sprintf("%s -%s- ", msg, strings.Repeat("-", pad))
}
