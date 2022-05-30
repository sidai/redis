package debug

import (
	"fmt"
	"net"
	"strings"

	"github.com/go-redis/redis/v8/internal/proto"
)

func ParseError(err error) string {
	switch e := err.(type) {
	case *net.OpError:
		return fmt.Sprintf("net-%s", e.Err)
	case net.Error:
		return fmt.Sprintf("net-%s", e)
	case proto.RedisError:
		switch errString := e.Error(); {
		case strings.HasPrefix(errString, "LOADING"):
			return "redis-LOADING"
		case strings.HasPrefix(errString, "MOVED "):
			return "redis-MOVED"
		case strings.HasPrefix(errString, "ASK "):
			return "redis-ASK"
		default:
			return fmt.Sprintf("redis-%s", strings.TrimPrefix(errString, "redis: "))
		}
	default:
		return e.Error()
	}
}
