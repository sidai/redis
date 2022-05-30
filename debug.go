package redis

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8/debug"
)

type ClusterSlots []ClusterSlot

func (s ClusterSlots) StatTags() [][]string {
	var tagsList [][]string
	for _, slot := range s {
		slotTag := fmt.Sprintf("slot:%d-%d", slot.Start, slot.End)
		for i, node := range slot.Nodes {
			addrTag := fmt.Sprintf("addr:%s", node.Addr)
			if i == 0 { // master
				tagsList = append(tagsList, []string{"type:master", slotTag, addrTag})
			} else { // slave
				tagsList = append(tagsList, []string{"type:slave", slotTag, addrTag})
			}
		}
	}
	return tagsList
}

func (s ClusterSlots) String() string {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Start < s[j].Start
	})
	var slots []string
	for _, slot := range s {
		slots = append(slots, slot.String())
	}

	return strings.Join(slots, "\n")
}

func (s ClusterSlot) String() string {
	var addrs []string
	for _, node := range s.Nodes {
		addrs = append(addrs, hostString(node.Addr))
	}

	return fmt.Sprintf("%-5d - %-5d: [%s]", s.Start, s.End, strings.Join(addrs, ", "))
}

func (n *clusterNode) IP() string {
	if n.Client == nil {
		return ""
	}
	return hostString(n.Client.opt.Addr)
}

func hostString(addr string) string {
	host, _, _ := net.SplitHostPort(addr)
	return host
}

var id uint32

func (c *clusterNodes) report(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ctx := debug.NewDebugCtx(ctx, fmt.Sprintf("%d-RNodes", atomic.AddUint32(&id, 1)))
		actives := make(map[string]bool)
		for _, addr := range c.activeAddrs {
			actives[addr] = true
		}

		var nodes []string
		for _, addr := range c.addrs {
			latency := int64(-1)
			if client, ok := c.nodes[addr]; ok {
				latency = client.Latency().Milliseconds()
			}
			nodes = append(nodes, fmt.Sprintf("addr: %s, active: %t, latency: %dms", hostString(addr), actives[addr], latency))
		}
		debug.Debugf(ctx, "Cluster Nodes Summary:\n%s", debug.Padding(strings.Join(nodes, "\n"), 8))
	}
}
