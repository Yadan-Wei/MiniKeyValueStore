package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
	maxPool             = 3
	snapshotInterval    = 20 * time.Second
)

type RaftNode struct {
	raft     *raft.Raft // raft client
	fsm      *FSM       // fsm of raft log
	leaderCh chan bool  // if node is a master
}

// TCP communication between raft nodes
func newRaftTransport(opts *Options) (*raft.NetworkTransport, error) {
	addr, err := net.ResolveTCPAddr("tcp", opts.raftTCPAddress)

	if err != nil {
		return nil, err
	}
	// func NewTCPTransport(bindAddr string,advertise net.Addr,maxPool int,
	// 	timeout time.Duration,logOutput io.Writer, ) (*NetworkTransport, error)
	trans, err := raft.NewTCPTransport(addr.String(), addr, maxPool, raftTimeout, os.Stderr)

	if err != nil {
		return nil, err
	}
	return trans, nil
}

// func NewRaft(conf *Config, fsm FSM, logs LogStore, stable StableStore,
// snaps SnapshotStore, trans Transport) (*Raft, error)
func newRaftNode(opts *Options, ctx *KVStoreContext) (*RaftNode, error) {
	raftConfig := raft.DefaultConfig()
	// raft node id
	raftConfig.LocalID = raft.ServerID(opts.raftTCPAddress)
	raftConfig.SnapshotInterval = snapshotInterval     // snapshort interval
	raftConfig.SnapshotThreshold = retainSnapshotCount // when more than 2 new log, do snapshot
	leaderCh := make(chan bool, 1)
	raftConfig.NotifyCh = leaderCh

	// network communication
	trans, err := newRaftTransport(opts)
	if err != nil {
		return nil, err
	}

	// create disk direction
	if err := os.MkdirAll(opts.dataDir, 0700); err != nil {
		return nil, err
	}

	//fsm
	fsm := &FSM{
		ctx: ctx,
		log: log.New(os.Stderr, "FSM: ", log.Ldate|log.Ltime)}

	// snapshot store
	snapshotStore, err := raft.NewFileSnapshotStore(opts.dataDir, 1, os.Stderr)
	if err != nil {
		return nil, err
	}

	// log store, here's embedded db store
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(opts.dataDir, "raftNode-log.bolt"))
	if err != nil {
		return nil, err
	}

	// kv value store, here's embedded db store
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(opts.dataDir, "raftNode-stable.bolt"))
	if err != nil {
		return nil, err
	}
	// raft node
	raftNode, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, trans)
	if err != nil {
		return nil, err
	}

	// if raft node is master, initialize the raft cluster
	if opts.bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: trans.LocalAddr(),
				},
			},
		}
		raftNode.BootstrapCluster(configuration)
	}

	return &RaftNode{raft: raftNode, fsm: fsm, leaderCh: leaderCh}, nil

}

// add a raft node to cluster
func joinRaftCluster(opts *Options) error {
	url := fmt.Sprintf("http://%s/join?peerAddress=%s", opts.joinAddress, opts.raftTCPAddress)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if string(body) != "ok" {
		return errors.New(fmt.Sprintf("Error joining cluster: %s", body))
	}

	return nil
}
