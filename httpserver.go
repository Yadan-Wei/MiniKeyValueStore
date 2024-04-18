package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/hashicorp/raft"
)

const (
	WRITE_DISABLED int32 = 0
	WRITE_ENABLED  int32 = 1
)

// http server

type HttpServer struct {
	ctx *KVStoreContext
	log *log.Logger
	mux *http.ServeMux
	// only master node can accept write request
	write int32
}

func NewHttpServer(ctx *KVStoreContext, log *log.Logger) *HttpServer {
	mux := http.NewServeMux()
	s := &HttpServer{
		ctx:   ctx,
		log:   log,
		mux:   mux,
		write: WRITE_DISABLED,
	}

	mux.HandleFunc("/set", s.doSet)
	mux.HandleFunc("/get", s.doGet)
	mux.HandleFunc("/join", s.doJoin)
	return s
}

func (hs *HttpServer) canWrite() bool {
	return atomic.LoadInt32(&hs.write) == WRITE_ENABLED
}

func (hs *HttpServer) setWriteFlag(flag bool) {
	if flag {
		atomic.StoreInt32(&hs.write, WRITE_ENABLED)
	} else {
		atomic.StoreInt32(&hs.write, WRITE_DISABLED)
	}
}

func (hs *HttpServer) doGet(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()

	key := vars.Get("key")

	if key == "" {
		hs.log.Println("Key cannot be empty, please provide a valid key")
		fmt.Fprint(w, "")
		return
	}

	// read value
	ret := hs.ctx.store.kv.Get(key)
	fmt.Fprintf(w, "%s\n", ret)
}

func (hs *HttpServer) doSet(w http.ResponseWriter, r *http.Request) {
	if !hs.canWrite() {
		fmt.Fprint(w, "write method not allowed\n")
		return
	}
	vars := r.URL.Query()

	key := vars.Get("key")
	value := vars.Get("value")
	if key == "" || value == "" {
		hs.log.Println("get nil key or nil value")
		fmt.Fprint(w, "param error\n")
		return
	}

	// change key value to log and serilization
	event := LogEntry{Key: key, Value: value}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		hs.log.Printf("json.Marshal failed, err:%v", err)
		fmt.Fprint(w, "internal error\n")
		return
	}

	// submit log to raft nodes
	applyFuture := hs.ctx.store.raftNode.raft.Apply(eventBytes, 5*time.Second)
	// submit log fail
	if err := applyFuture.Error(); err != nil {
		hs.log.Printf("raftNode.Apply failed:%v", err)
		fmt.Fprint(w, "internal error\n")
		return
	}
	// submit log success
	fmt.Fprintf(w, "ok\n")
}

// doJoin handles joining cluster request
func (hs *HttpServer) doJoin(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()

	peerAddress := vars.Get("peerAddress")
	if peerAddress == "" {
		hs.log.Println("invalid PeerAddress")
		fmt.Fprint(w, "invalid peerAddress\n")
		return
	}
	// add voter node in the cluster
	addPeerFuture := hs.ctx.store.raftNode.raft.AddVoter(raft.ServerID(peerAddress), raft.ServerAddress(peerAddress), 0, 0)
	if err := addPeerFuture.Error(); err != nil {
		hs.log.Printf("Error joining peer to raftNode, peeraddress:%s, err:%v, code:%d", peerAddress, err, http.StatusInternalServerError)
		fmt.Fprint(w, "internal error\n")
		return
	}
	fmt.Fprint(w, "ok")
}
