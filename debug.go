package redis

import (
	"fmt"
	"net"
	"sort"
	"strings"
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
