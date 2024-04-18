package main

import (
	"encoding/json"
	"io"
	"log"
	"sync"
)

// the KVstore struct
type KV struct {
	data map[string]string
	sync.RWMutex
}

// init a kv store
func NewKV() *KV {
	kv := &KV{}
	kv.data = make(map[string]string)
	return kv
}

// set key value pair
func (kv *KV) Set(key string, value string) error {
	kv.Lock()
	defer kv.Unlock()
	kv.data[key] = value
	return nil
}

// retrive value from kv store
func (kv *KV) Get(key string) string {
	kv.RLock()
	ret := kv.data[key]
	kv.RUnlock()
	return ret
}

// seriallize key value data
func (kv *KV) Marshal() ([]byte, error) {
	kv.RLock()
	defer kv.RUnlock()
	dataStreams, err := json.Marshal(kv.data)
	return dataStreams, err
}

func (kv *KV) UnMarshal(serialized io.ReadCloser) error {
	var newData map[string]string
	if err := json.NewDecoder(serialized).Decode(&newData); err != nil {
		return err
	}
	kv.Lock()
	defer kv.Unlock()
	kv.data = newData
	return nil
}

type KVStore struct {
	httpServer *HttpServer // http server
	opts       *Options    // configuration
	log        *log.Logger
	kv         *KV       // kvCache server
	raftNode   *RaftNode // raft node
}

// KVStoreContext
type KVStoreContext struct {
	store *KVStore
}
