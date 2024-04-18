package main

import (
	"encoding/json"
	"io"
	"log"

	"github.com/hashicorp/raft"
)

// Implement FSM interface of raft
type FSM struct {
	ctx *KVStoreContext
	log *log.Logger
}

// KV log data
type LogEntry struct {
	Key   string
	Value string
}

func (fsm *FSM) Apply(logEntry *raft.Log) interface{} {
	var e LogEntry // If unsure of type, use interface{}
	if err := json.Unmarshal(logEntry.Data, &e); err != nil {
		panic("Cannot unmarshal Raft log: " + err.Error())
	}

	ret := fsm.ctx.store.kv.Set(e.Key, e.Value)

	fsm.log.Printf("FSM.Apply(), logEntry:%s, ret:%v\n", logEntry.Data, ret)
	return ret
}

// Snapshot returns a latest Snapshot
func (fsm *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &Snapshot{cm: fsm.ctx.store.kv}, nil
}

// Restore the key-value store to a previous state.
func (fsm *FSM) Restore(serialized io.ReadCloser) error {
	return fsm.ctx.store.kv.UnMarshal(serialized)
}
