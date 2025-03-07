package repo

import (
	"fmt"
	"go-redis/pkg/utils"
	"strconv"
	"sync"
)

const (
	MAXSIZE = 2
)

type List struct {
	KvList map[string]*QuickList
	mu     sync.Mutex
}

type Node struct {
	data []string
	next *Node
	prev *Node
}
type QuickList struct {
	head    *Node
	tail    *Node
	maxSize int // max size of the Node : how many ziplist can be stored in a Node
	length  int // total number
	mu      sync.Mutex
}

var KvList *List

func NewKvList() *List {
	return &List{
		KvList: make(map[string]*QuickList),
	}
}

func InitKVList() {
	KvList = NewKvList()
}

func NewQuickList() *QuickList {
	return &QuickList{
		maxSize: MAXSIZE,
	}
}

func NewNode() *Node {
	return &Node{
		//  len set 0 , slice will not be fill by ""
		//  it will affect append later
		data: make([]string, MAXSIZE),
	}
}

func (l *List) GetQuickList(key string) *QuickList {
	if _, ok := l.KvList[key]; !ok {
		l.KvList[key] = NewQuickList()
	}
	return l.KvList[key]
}

func (l *List) Lpush(key, value string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	ql := l.GetQuickList(key)
	if ql.head == nil {
		ql.head = NewNode()
		ql.tail = ql.head
	}
	if ql.head.data[len(ql.head.data)-1] != "" {
		newNode := NewNode()
		newNode.next = ql.head
		ql.head.prev = newNode
		ql.head = newNode
	}
	//  put new data to node head
	newSlice := make([]string, MAXSIZE)
	newSlice[0] = value
	copy(newSlice[1:], ql.head.data)
	ql.head.data = newSlice
	ql.length++
	return nil
}
func (l *List) Rpush(key, value string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	ql := l.GetQuickList(key)
	if ql.head == nil {
		ql.head = NewNode()
		ql.tail = ql.head
	}
	if len(ql.head.data) >= ql.maxSize {
		newNode := NewNode()
		newNode.next = ql.head
		ql.head.prev = newNode
		ql.head = newNode
	}
	//  put new data to node head
	ql.head.data = append(ql.head.data, value)
	ql.length++
	return nil
}

func (l *List) Lrange(key, start, end string) ([]string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	ql := l.GetQuickList(key)
	err, s, e := isRangeValid(ql, start, end)
	if err != nil {
		return nil, err
	}
	if e == 0 && s == 0 {
		return []string{}, nil
	}
	res := ql.iterQuickList(s, e)
	return res, nil
}

func (ql *QuickList) iterQuickList(start, end int) []string {
	var res []string
	index := 0
	if start == -1 {
		for node := ql.head; node != nil; node = node.next {
			res = append(res, node.data...)
		}
		return res
	}
	for node := ql.head; node != nil; node = node.next {
		dataLen := len(node.data)
		nodeEnd := index + dataLen - 1
		// if not in current node
		if nodeEnd < start {
			index += dataLen
			continue
		}
		// if start end is not in same nodedata
		relativeStart := max(0, start-index)
		relativeEnd := min(dataLen, end-index+1)
		res = append(res, node.data[relativeStart:relativeEnd]...)
		index += dataLen
		//  find end , break the loop
		if end < nodeEnd {
			break
		}
		// low performance
		// for _, v := range node.data {
		// 	if index >= start && index <= end {
		// 		res = append(res, v)
		// 	}
		// 	index++
		// 	if index > end {
		// 		return res
		// 	}
		// }

	}
	return res
}

func isRangeValid(ql *QuickList, start, end string) (error, int, int) {
	if !utils.IsNumeric(start) || !utils.IsNumeric(end) {
		return fmt.Errorf("range must be numeric"), 0, 0
	}
	s, _ := strconv.Atoi(start)
	e, _ := strconv.Atoi(end)
	if s < -1 || e < -1 {
		return fmt.Errorf("range must be bigger than 0"), 0, 0
	}
	if s > e {
		return nil, 0, 0 // 返回空范围
	}
	if s >= ql.length {
		return nil, 0, 0 // `start` 超出范围，返回空
	}
	if e >= ql.length {
		e = ql.length - 1 // 限制 `end` 到最大索引
	}
	return nil, s, e
}
