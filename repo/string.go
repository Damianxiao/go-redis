package repo

import (
	"fmt"
	"go-redis/pkg/utils"
	"strconv"
	"sync"
	"time"
)

var KvMemory KV

type KV struct {
	kv       map[string][]byte
	kvExpire map[string]time.Time
	mu       sync.Mutex
}

func NewKV() KV {
	return KV{
		kv:       make(map[string][]byte),
		kvExpire: make(map[string]time.Time),
	}
}

func InitKV() {
	KvMemory = NewKV()
}

func (kv KV) Set(key, val string, ex ...string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.kv[key] = []byte(val)
	// set time limit
	now := time.Now()
	num := 0
	if len(ex) != 0 {
		if ex[0] != "" {
			num, _ = strconv.Atoi(ex[0])
			exp := now.Add(time.Duration(num) * time.Millisecond)
			kv.kvExpire[key] = exp
		}
	}
	return nil
}

func (kv KV) Get(key string) ([]byte, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	val, ok := kv.kv[key]
	if exp, exist := kv.kvExpire[key]; exist && ok {
		if time.Since(exp) > 0 {
			delete(kv.kv, key)
			return nil, fmt.Errorf("data expired already")
		}
	}
	if ok {
		return val, nil
	}
	return nil, fmt.Errorf("data not exist")
}

func (kv KV) Delete(key string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	delete(kv.kv, key)
	delete(kv.kvExpire, key)
	return nil
}

func (kv KV) Exist(key string) (bool, error) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	_, ok := kv.kv[key]
	if ok {
		return true, nil
	}
	return false, nil
}

func (kv KV) Incr(key string, amount ...string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	plus := 1
	if len(amount) != 0 {
		if amount[0] != "" {
			plus, _ = strconv.Atoi(amount[0])
		}
	}
	if val, ok := kv.kv[key]; ok && utils.IsNumeric(string(val)) {
		now, _ := strconv.Atoi(string(val))
		v := []byte(fmt.Sprintf("%d", plus+now))
		kv.kv[key] = v
	}
	return nil
}

func (kv KV) Decr(key string, amount ...string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	plus := -1
	if len(amount) != 0 {
		if amount[0] != "" {
			plus, _ = strconv.Atoi(amount[0])
			plus *= -1
		}
	}
	if val, ok := kv.kv[key]; ok && utils.IsNumeric(string(val)) {
		now, _ := strconv.Atoi(string(val))
		v := []byte(fmt.Sprintf("%d", plus+now))
		kv.kv[key] = v
	}
	return nil
}
