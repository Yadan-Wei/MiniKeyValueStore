package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	store := &KVStore{
		opts: NewOptions(),
		log:  log.New(os.Stderr, "KVStore: ", log.Ldate|log.Ltime),
		kv:   NewKV(),
	}
	ctx := &KVStoreContext{store}

	l, err := net.Listen("tcp", store.opts.httpAddress)
	if err != nil {
		store.log.Fatal(fmt.Sprintf("listen %s failed: %s", store.opts.httpAddress, err))
	}
	store.log.Printf("http server listen:%s", l.Addr())

	// new goroutine to run httpserver
	logger := log.New(os.Stderr, "httpserver: ", log.Ldate|log.Ltime)
	httpServer := NewHttpServer(ctx, logger)
	store.httpServer = httpServer
	go func() {
		_ = http.Serve(l, httpServer.mux)
	}()

	// new RaftNode

	raft, err := newRaftNode(store.opts, ctx)

	if err != nil {
		store.log.Fatal(fmt.Sprintf("new raftNode node failed:%v", err))
	}

	store.raftNode = raft

	// Non-leader raft node join the raft cluster
	if store.opts.joinAddress != "" {
		err = joinRaftCluster(store.opts)
		if err != nil {
			store.log.Fatal(fmt.Sprintf("join raftNode cluster failed:%v", err))
		}
	}

	// main goroutine monitor leader election, set self state
	for {
		select {
		case leader := <-store.raftNode.leaderCh:
			if leader {
				store.log.Println("become leader, enable write api")
				store.httpServer.setWriteFlag(true)
			} else {
				store.log.Println("become follower, close write api")
				store.httpServer.setWriteFlag(false)
			}
		}
	}

}
