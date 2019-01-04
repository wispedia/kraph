package kraph

import (
	"fmt"
	"github.com/pquerna/ffjson/ffjson"
	"sync"
)

type ID interface {
	String() string
}

type nid string

func (id nid) String() string {
	return string(id)
}

func NewNID(id string) ID {
	return nid(id)
}

// node definition
type Node interface {
	GetId() ID
}

type node struct {
	id ID
}

func (n *node) GetId() ID {
	return n.id
}

func NewNode(id ID) Node {
	return &node{
		id: id,
	}
}

// Graph definition
type Graph interface {
	// 重置 graph ，会删除其中所有的边和节点
	Init()

	// 返回 graph 中所有节点的数量
	GetNodeCount() int

	// 通过 id 在图中查找节点，如果节点不存在，则会返回 nil
	GetNode(id ID) Node

	// 返回 graph 中所有node
	GetNodes() map[ID]Node

	// 向图中添加 node 如果该 node 已经存在则返回 false
	AddNode(nd Node) bool

	// 从图中删除 node 如果 node 不存在，则返回 false
	DeleteNode(id ID) bool

	// 将图中的两个 node 建立关系，并增加权重，如果 node 不存在则返回 error
	// 如果两个 node 已经存在关系，则权重相加
	AddEdge(id, pid ID, wgt float64) error

	// 替换两个 node 之间的权重，如果 node 不存在则返回 error
	ReplaceEdge(id, pid ID, wgt float64) error

	// 删除两个 node 之间的关系，如果 node 不存在则返回 error
	DeleteEdge(id, pid ID) error

	// 获取两个 node 之间的权重
	GetWeight(id, pid ID) (float64, error)

	// 获取给定 node 的所有上游
	GetSources(id ID) (map[ID]Node, error)

	// 获取给定 node 的所有下游
	GetTargets(id ID) (map[ID]Node, error)

	// 将整个图输出为 json 格式
	JSON() ([]byte, error)
}

func NewGraph() Graph {
	return &graph{
		nodeList:    make(map[ID]Node),
		nodeSources: make(map[ID]map[ID]float64),
		nodeTargets: make(map[ID]map[ID]float64),
	}
}

type graph struct {
	mu          sync.RWMutex
	nodeList    map[ID]Node
	nodeSources map[ID]map[ID]float64
	nodeTargets map[ID]map[ID]float64
}

func (g *graph) Init() {
	g.nodeList = make(map[ID]Node)
	g.nodeSources = make(map[ID]map[ID]float64)
	g.nodeTargets = make(map[ID]map[ID]float64)
}

func (g *graph) GetNodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.nodeList)
}

func (g *graph) GetNode(id ID) Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.nodeList[id]
}

func (g *graph) GetNodes() map[ID]Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.nodeList
}

func (g *graph) unsafeIdExist(id ID) bool {
	_, ok := g.nodeList[id]

	return ok
}

func (g *graph) AddNode(nd Node) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 如果这个节点已经存在，返回false
	if g.unsafeIdExist(nd.GetId()) {
		return false
	}

	id := nd.GetId()
	g.nodeList[id] = nd

	return true
}

func (g *graph) DeleteNode(id ID) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 如果这个 id 在 node list 中不存在 直接返回false
	if !g.unsafeIdExist(id) {
		return false
	}
	delete(g.nodeList, id)
	delete(g.nodeTargets, id)

	for _, tmap := range g.nodeTargets {
		delete(tmap, id)
	}

	delete(g.nodeSources, id)

	for _, smap := range g.nodeSources {
		delete(smap, id)
	}

	return true
}

func (g *graph) AddEdge(id, pid ID, wgt float64) error {
	// 如果已经存在此条关系，则增加其权重，如果没有则创建
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.unsafeIdExist(id) {
		return fmt.Errorf("%s does not exist in graph", id)
	}

	if !g.unsafeIdExist(pid) {
		return fmt.Errorf("%s does not exist in graph", pid)
	}

	if _, ok := g.nodeTargets[pid]; ok {
		if w, ok2 := g.nodeTargets[pid][id]; ok2 {
			g.nodeTargets[pid][id] = w + wgt
		} else {
			g.nodeTargets[pid][id] = wgt
		}
	} else {
		g.nodeTargets[pid] = map[ID]float64{
			id: wgt,
		}
	}

	if _, ok := g.nodeSources[id]; ok {
		if w, ok2 := g.nodeSources[id][pid]; ok2 {
			g.nodeSources[id][pid] = w + wgt
		} else {
			g.nodeSources[id][pid] = wgt
		}

	} else {
		g.nodeSources[id] = map[ID]float64{
			pid: wgt,
		}
	}

	return nil
}

func (g *graph) ReplaceEdge(id, pid ID, wgt float64) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.unsafeIdExist(id) {
		return fmt.Errorf("%s does not exist in graph", id)
	}

	if !g.unsafeIdExist(pid) {
		return fmt.Errorf("%s does not exist in graph", pid)
	}

	if _, ok := g.nodeTargets[pid]; ok {
		g.nodeTargets[pid][id] = wgt
	} else {
		g.nodeTargets[pid] = map[ID]float64{
			id: wgt,
		}
	}

	if _, ok := g.nodeSources[id]; ok {
		g.nodeSources[id][pid] = wgt
	} else {
		g.nodeSources[id] = map[ID]float64{
			pid: wgt,
		}
	}

	return nil

}

func (g *graph) DeleteEdge(id, pid ID) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.unsafeIdExist(id) {
		return fmt.Errorf("%s does not exist in graph", id)
	}

	if !g.unsafeIdExist(pid) {
		return fmt.Errorf("%s does not exist in graph", pid)
	}

	if _, ok := g.nodeTargets[pid]; ok {
		if _, ok := g.nodeTargets[pid][id]; ok {
			delete(g.nodeTargets[pid], id)
		}
	}

	if _, ok := g.nodeSources[id]; ok {
		if _, ok := g.nodeSources[id][pid]; ok {
			delete(g.nodeSources[id], pid)
		}
	}

	return nil
}

func (g *graph) GetWeight(id, pid ID) (float64, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.unsafeIdExist(id) {
		return 0.0, fmt.Errorf("%s does not exist in graph", id)
	}

	if !g.unsafeIdExist(pid) {
		return 0.0, fmt.Errorf("%s does not exist in graph", pid)
	}

	if _, ok := g.nodeSources[id]; ok {
		if w, ok := g.nodeSources[id][pid]; ok {
			return w, nil
		}
	}

	return 0.0, fmt.Errorf("there is no edge from %s to %s", pid, id)

}

func (g *graph) GetSources(id ID) (map[ID]Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.unsafeIdExist(id) {
		return nil, fmt.Errorf("%s does not exist in graph", id)
	}

	s := make(map[ID]Node)

	if _, ok := g.nodeSources[id]; ok {
		for pid := range g.nodeSources[id] {
			s[pid] = g.nodeList[pid]
		}
	}

	return s, nil
}

func (g *graph) GetTargets(pid ID) (map[ID]Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.unsafeIdExist(pid) {
		return nil, fmt.Errorf("%s does not exist in graph", pid)
	}

	t := make(map[ID]Node)

	if _, ok := g.nodeTargets[pid]; ok {
		for pid := range g.nodeTargets[pid] {
			t[pid] = g.nodeList[pid]
		}
	}

	return t, nil
}

func (g *graph) JSON() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	rs := make(map[string]map[string]float64)
	for id, _ := range g.nodeList {
		smap, err := g.GetSources(id)
		if err != nil {
			continue
		}

		for pid, _ := range smap {
			if wgt, err := g.GetWeight(id, pid); err == nil {
				if _, ok := rs[id.String()]; ok {
					rs[id.String()][id.String()] = wgt
				} else {
					rs[id.String()] = map[string]float64{pid.String(): wgt}
				}
			}
		}
	}
	return ffjson.Marshal(rs)

}
