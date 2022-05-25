package internal

import (
	"fmt"
	"net"
	"strings"
)

func ParseError(err error) string {
	switch e := err.Error(); {
	case strings.HasPrefix(e, "LOADING"):
		return "redis-LOADING"
	case strings.HasPrefix(e, "MOVED "):
		return "redis-MOVED"
	case strings.HasPrefix(e, "ASK "):
		return "redis-ASK"
	case IsInternalError(err):
		return fmt.Sprintf("redis-%s", strings.TrimPrefix(e, "redis: "))
	case IsNetworkError(err):
		if opErr, ok := err.(*net.OpError); ok {
			return fmt.Sprintf("net-%s", opErr.Err)
		}
		return fmt.Sprintf("net-%s", e)
	default:
		return e
	}
}
