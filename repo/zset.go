package repo

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
)

const (
	maxHeight = 31
)

// Level 3:  A  ────────────>  E
// Level 2:  A  ───>  C  ───>  E
// Level 1:  A  ───>  C  ───>  D  ───>  E
// Level 0:  A  ─>  B  ─>  C  ─>  D  ─>  E
// 5
//
//	in the list, total node nums len(skiplist) : 5 + 2(head and tail)
//	as every node , his params {forward}[i] represent the next node in the i level,
//
// A := &SkipListNode{
// member: "A", score: 1.0,
// forward: []*SkipListNode{B, C, C, E},
// }
type SkipListNode struct {
	member  string
	score   float64
	forward []*SkipListNode // next floor
}

type SkipList struct {
	head   *SkipListNode
	height int
	length int
}

type Zset struct {
	dict     map[string]*SkipListNode
	skiplist *SkipList
}

type KvZset struct {
	Zset map[string]*Zset
	mu   sync.Mutex
}

var MemoryZset *KvZset

func NewMemoryZset() *KvZset {
	return &KvZset{
		Zset: make(map[string]*Zset),
	}
}

func InitKvZset() {
	MemoryZset = NewMemoryZset()
}

func NewSkipListNode(member string, score float64) *SkipListNode {
	return &SkipListNode{
		member:  member,
		score:   score,
		forward: make([]*SkipListNode, maxHeight),
	}
}

func NewSkipList() *SkipList {
	return &SkipList{
		head:   NewSkipListNode("", math.Inf(-1)),
		height: 1,
		length: 0,
	}
}

func NewZset() *Zset {
	return &Zset{
		dict:     make(map[string]*SkipListNode),
		skiplist: NewSkipList(),
	}
}

func (kz *KvZset) GetZset(key string) *Zset {
	if _, ok := kz.Zset[key]; !ok {
		kz.Zset[key] = NewZset()
	}
	return kz.Zset[key]
}

func (kz *KvZset) Insert(key, member string, score float64) error {
	kz.mu.Lock()
	defer kz.mu.Unlock()
	// update represent insert pos
	update := make([]*SkipListNode, maxHeight)
	zset := kz.GetZset(key)

	if zset.skiplist.head == nil {
		zset.skiplist.head = NewSkipListNode("", math.Inf(-1))
	}
	list := zset.skiplist
	curr := list.head
	// find insert spot pre node
	for i := list.height - 1; i > 0; i-- {
		for curr.forward[i] != nil {
			if curr.forward[i].score < score {
				curr = curr.forward[i]
			}
		}
		update[i] = curr

	}
	// generate level
	level := randomLevel()
	if level > list.height {
		for i := list.height - 1; i <= level; i++ {
			update[i] = list.head
		}
		list.height = level
	}

	// insert and change forward
	newNode := NewSkipListNode(member, score)
	for i := list.height - 1; i > 0; i-- {
		if update[i].forward[i] != nil {
			newNode.forward[i] = update[i].forward[i]
			update[i].forward[i] = newNode
		}
		newNode.forward[i] = nil
		update[i].forward[i] = newNode
	}
	zset.dict[member] = newNode
	// add length of list
	list.length++
	return nil
}

func (kv *KvZset) GetScore(key string, value string) float64 {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	return kv.Zset[key].dict[value].score
}

func (kv *KvZset) Zrank(key string, member string) (error, int) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	zset := kv.GetZset(key)
	var memberScore float64
	if _, ok := zset.dict[member]; !ok {
		return fmt.Errorf("member is not exist"), -1
	}
	memberScore = zset.dict[member].score
	list := zset.skiplist
	curr := zset.skiplist.head
	rank := 1
	for i := list.height - 1; i >= 0; i-- {
		for curr.forward[i] != nil {
			if curr.forward[i].score < memberScore {
				rank++
				curr = curr.forward[i]
			} else if curr.forward[i].score == memberScore {
				return nil, rank
			}
		}
	}
	return nil, 0
}

func (kz *KvZset) Zrm(key, member string) error {
	zset := kz.GetZset(key)
	if _, ok := zset.dict[member]; !ok {
		return fmt.Errorf("member not exist")
	}
	list := zset.skiplist
	curr := list.head
	update := make([]*SkipListNode, list.height)
	// make sure level 0 be delete
	for i := list.height - 1; i >= 0; i-- {
		for curr.forward[i] != nil &&
			(curr.forward[i].score < zset.dict[member].score ||
				(curr.forward[i].score == zset.dict[member].score &&
					curr.forward[i].member == member)) {
			curr = curr.forward[i]
		}
		update[i] = curr
	}
	target := curr.forward[0]

	for i := range list.height {
		if update[i].forward[i] != target {
			break
		}
		update[i].forward[i] = target.forward[i]
	}

	for i, v := range curr.forward {
		if v.member == member {
			curr.forward[i] = curr.forward[i].forward[i]
		}
	}

	for list.height > 1 && list.head.forward[list.height-1] == nil {
		list.height--
	}

	delete(zset.dict, member)
	list.length--
	return nil
}
func randomLevel() int {
	return rand.Intn(maxHeight)
}
