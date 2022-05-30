package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis/v8/cmd/config"
	"github.com/go-redis/redis/v8/debug"
)

const (
	intervalSeconds = 10
)

var (
	debugger = debug.FromPath("redis-cluster.txt")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	debug.SetDebugger(debugger)
	statHostLookup(ctx, config.RedisHost)
	statClusterSlots(ctx, config.RedisHost, config.RedisPort)

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
}

func tickerRun(fn func()) {
	go func() {
		ticker := time.NewTicker(time.Second * intervalSeconds)
		for range ticker.C {
			fn()
		}
	}()
}

func statHostLookup(ctx context.Context, dns string) {
	tickerRun(func() {
		ctx = debug.NewDebugCtx(ctx, "LookUp")
		addrs, err := net.LookupHost(dns)
		if err != nil {
			debug.Debugf(ctx, "Unable to lookup host %s: %+v", dns, err)
			return
		}

		debug.Debugf(ctx, "Lookup %s: [%+v]", dns, strings.Join(addrs, ", "))
	})
}

func statClusterSlots(ctx context.Context, host string, port string) {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
	})
	tickerRun(func() {
		ctx = debug.NewDebugCtx(ctx, "LoadSlots")
		slots, err := rdb.ClusterSlots(ctx).Result()
		if err != nil {
			debug.Debugf(ctx, "Unable to load cluster slots: %s", err)
			return
		}

		debug.Debugf(ctx, "Cluster Slots Summary:\n%s", debug.Padding(redis.ClusterSlots(slots).String(), 8))
	})
}
