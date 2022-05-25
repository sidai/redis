package debug

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

var datadog = struct {
	client   statsd.ClientInterface
	debugger Debugging
}{
	client:   &statsd.NoOpClient{},
	debugger: &NoopDebugger{},
}

func Init(debugger Debugging, url string, tags ...string) {
	client, err := statsd.New(url, statsd.WithTags(tags))
	if err != nil {
		panic(err)
	}

	datadog.client = client
	datadog.debugger = debugger
}

func Incr(name string, tags ...string) {
	IncrCtx(context.Background(), name, tags...)
}

func IncrCtx(ctx context.Context, name string, tags ...string) {
	go func() {
		err := datadog.client.Incr(name, tags, 1)
		if err != nil {
			datadog.debugger.Debugf(ctx, "metric: %s, unexpected error %s during incr with tags %+v", name, err, tags)
		}
	}()
}

func Count(name string, value int64, tags ...string) {
	CountCtx(context.Background(), name, value, tags...)
}

func CountCtx(ctx context.Context, name string, value int64, tags ...string) {
	go func() {
		err := datadog.client.Count(name, value, tags, 1)
		if err != nil {
			datadog.debugger.Debugf(ctx, "metric: %s, unexpected error %s during count with tags %+v", name, err, tags)
		}
	}()
}

func Gauge(name string, value float64, tags ...string) {
	GaugeCtx(context.Background(), name, value, tags...)
}

func GaugeCtx(ctx context.Context, name string, value float64, tags ...string) {
	go func() {
		err := datadog.client.Gauge(name, value, tags, 1)
		if err != nil {
			datadog.debugger.Debugf(ctx, "metric: %s, unexpected error %s during gauge with tags %+v", name, err, tags)
		}
	}()
}

func Enum(name string, tagsList ...[]string) {
	EnumCtx(context.Background(), name, tagsList...)
}

// EnumCtx generate the enum value by using the gauge operator with the current timestamp
func EnumCtx(ctx context.Context, name string, tagsList ...[]string) {
	go func() {
		var firstErr error
		timeTag := fmt.Sprintf("time:%s", time.Now().Format("04:05.000"))
		for _, newTags := range tagsList {
			err := datadog.client.Gauge(name, 1, append(newTags, timeTag), 1)
			if firstErr != nil && err != nil {
				firstErr = err
			}
		}

		if firstErr != nil {
			datadog.debugger.Debugf(ctx, "metric: %s, unexpected error %s during enum operation", name, firstErr)
		}
	}()
}

func Histogram(name string, start time.Time, tags ...string) {
	HistogramCtx(context.Background(), name, start, tags...)
}

func HistogramCtx(ctx context.Context, name string, start time.Time, tags ...string) {
	go func() {
		delta := float64(time.Since(start).Milliseconds())
		err := datadog.client.Histogram(name, delta, tags, 1)
		if err != nil {
			datadog.debugger.Debugf(ctx, "metric: %s, unexpected error %s during histogram with tags %+v", name, err, tags)
		}
	}()
}
