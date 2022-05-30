package config

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	defaultWorker    = 500
	defaultRedisHost = "stg-spiderman-redis-3-2-4.redis.sreopstool.grabtaxi.io"
	defaultRedisPort = "6379"
	defaultPoolSize  = 50
)

var (
	HostName   = getHostName()
	RedisHost  = getEnv("REDIS_HOST", defaultRedisHost)
	RedisPort  = getEnv("REDIS_PORT", defaultRedisPort)
	WorkerSize = getIntEnv("WORKER_SIZE", defaultWorker)
	PoolSize   = getIntEnv("POOL_SIZE", defaultPoolSize)
	LogFile    = getEnv("LOG_FILE", fmt.Sprintf("v8-%s-w%d-p%d.txt", HostName, WorkerSize, PoolSize))
)

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func getIntEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		worker, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}
		return worker
	}

	return fallback
}

func getHostName() string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	respChan := make(chan string)

	go func() {
		resp, err := http.Get("http://169.254.169.254/latest/meta-data/instance-id")
		if err != nil {
			respChan <- ""
			return
		}

		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			respChan <- ""
			return
		}
		respChan <- string(bodyBytes)
	}()

	select {
	case host := <-respChan:
		if host != "" {
			return host
		}
	case <-ctx.Done():
	}

	hostname, _ := os.Hostname()
	return hostname
}
