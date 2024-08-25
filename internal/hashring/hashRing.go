package hashring

import (
	"hash/crc32"
	"slices"
	"sort"
	"strconv"
	"sync"
)

type HashRingT struct {
	mu            sync.Mutex
	nodes         []NodeT
	vnodesPerNode int
}

type NodeT struct {
	hash int
	name string
}

func NewHashRing(vnodesPerNode int) *HashRingT {
	return &HashRingT{
		vnodesPerNode: vnodesPerNode,
	}
}

func (h *HashRingT) AddNodes(nodeNames []string) {
	h.mu.Lock()
	for _, nodeName := range nodeNames {
		for i := 0; i < h.vnodesPerNode; i++ {
			vnode := nodeName + "#" + strconv.Itoa(i)
			hash := int(crc32.ChecksumIEEE([]byte(vnode)))
			h.nodes = append(h.nodes, NodeT{hash: hash, name: nodeName})
		}

		sort.Slice(h.nodes, func(i, j int) bool {
			return h.nodes[i].hash < h.nodes[j].hash
		})
	}
	h.mu.Unlock()
}

func (hr *HashRingT) InHashRing(nodeName string) (result bool) {
	hr.mu.Lock()
	for _, node := range hr.nodes {
		if node.name == nodeName {
			result = true
			break
		}
	}
	hr.mu.Unlock()

	return result
}

func (hr *HashRingT) GetNodes() (result []NodeT) {
	hr.mu.Lock()
	result = hr.nodes
	hr.mu.Unlock()

	return result
}

func (h *HashRingT) RemoveNodes(nodes []string) {
	h.mu.Lock()
	var newNodes []NodeT

	for _, node := range h.nodes {
		if !slices.Contains(nodes, node.name) {
			newNodes = append(newNodes, node)
		}
	}

	h.nodes = newNodes
	h.mu.Unlock()
}

func (h *HashRingT) GetNode(key string) (node string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.nodes) == 0 {
		return ""
	}

	hash := int(crc32.ChecksumIEEE([]byte(key)))

	idx := sort.Search(len(h.nodes), func(i int) bool {
		return h.nodes[i].hash >= hash
	})

	if idx == len(h.nodes) {
		idx = 0
	}

	node = h.nodes[idx].name

	return node
}
