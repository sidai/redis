package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"gopkg.in/redis.v5"
	"gopkg.in/redis.v5/cmd/config"
	"gopkg.in/redis.v5/debug"
)

const (
	intervalSeconds = 10

	elapsedMetricName   = "spiderman.app.get.elapsed"
	requestMetricName   = "spiderman.app.get.request"
	goroutineMetricName = "spiderman.app.goroutine.total"
)

var (
	debugger = debug.FromPath(config.LogFile)
	reporter = debug.NewReporter(debugger)
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reporter.Report(ctx, time.Second*intervalSeconds)
	debug.SetDebugger(debugger)

	debug.Init(debugger, "127.0.0.1:8125", "appname:spiderman-test", fmt.Sprintf("host:%s", config.HostName))

	host, port, poolSize, workerSize := config.RedisHost, config.RedisPort, config.PoolSize, config.WorkerSize

	debug.Debugf(ctx, "Addr: %s, PoolSize: %d, WorkerSize: %d", host, poolSize, workerSize)

	statGoroutines(ctx)
	runRedisTest(ctx, []string{fmt.Sprintf("%s:%s", host, port)}, poolSize, workerSize)
}

func statGoroutines(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Second * intervalSeconds)
		for range ticker.C {
			debug.GaugeCtx(ctx, goroutineMetricName, float64(runtime.NumGoroutine()))
		}
	}()
}

func runRedisTest(ctx context.Context, addrs []string, poolSize int, workerSize int) {
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:          addrs,
		ReadOnly:       true,
		RouteByLatency: true,
		PoolSize:       poolSize,
	})

	populateDatabase(ctx, rdb, 10000)

	runTest(ctx, rdb, 10000, workerSize)
	quitIfInterrupt(func() {
		cleanDatabase(ctx, rdb)
	})
}

func runTest(ctx context.Context, rdb redis.Cmdable, keyLength int, worker int) {
	for i := 0; i < worker; i++ {
		go func() {
			for {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)+100)) // sleep around 100ms - 299ms, i.e, QPS: ave 5 per worker, max 10 per worker
				go func(ctx context.Context) {
					key := strconv.Itoa(rand.Intn(keyLength) + 1)
					start := time.Now()

					var err error
					var cmdTag string

					if rand.Intn(2) == 0 {
						ctx = debug.NewDebugCtx(ctx, fmt.Sprintf("Get-%s", key))
						cmd := rdb.Get(key)
						_, err = cmd.Result()
						reporter.AddStat(cmd.Stat())
						cmdTag = "cmd:get"
					} else {
						ctx = debug.NewDebugCtx(ctx, fmt.Sprintf("Set-%s", key))
						cmd := rdb.Set(key, key, time.Hour)
						_, err = cmd.Result()
						reporter.AddStat(cmd.Stat())
						cmdTag = "cmd:set"
					}

					if err != nil && err != redis.Nil {
						debug.IncrCtx(ctx, requestMetricName, cmdTag, "type:failed", fmt.Sprintf("reason:%s", err))
						debug.HistogramCtx(ctx, elapsedMetricName, start, cmdTag, "type:failed", fmt.Sprintf("reason:%s", err))
						reporter.Fail()
					} else {
						debug.IncrCtx(ctx, requestMetricName, cmdTag, "type:success")
						debug.HistogramCtx(ctx, elapsedMetricName, start, cmdTag, "type:success")
					}
					reporter.Done()
				}(ctx)
			}
		}()
	}
}

func populateDatabase(ctx context.Context, rdb redis.Cmdable, keyCount int) {
	ctx = debug.NewDebugCtx(ctx, "Populate")
	cmds, err := rdb.Pipelined(func(p *redis.Pipeline) error {
		for id := 1; id <= keyCount; id++ {
			p.Set(strconv.Itoa(id), strconv.Itoa(id), time.Hour)
		}
		return nil
	})

	failed := 0
	if err != nil {
		for _, cmd := range cmds {
			if err := cmd.Err(); err != nil && err != redis.Nil {
				failed++
			}
		}
	}

	debug.Debugf(ctx, "Written key value pair of size %+v into Redis cluster\n", keyCount-failed)
}

func cleanDatabase(ctx context.Context, rdb redis.Cmdable) {
	ctx = debug.NewDebugCtx(ctx, "CleanUp")
	rdb.FlushAll()

	debug.Debugf(ctx, "Flush Redis cluster completed\n")
}

func quitIfInterrupt(cleanup func()) {
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel

	cleanup()
}
