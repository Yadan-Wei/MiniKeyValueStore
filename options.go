package main

import (
	"flag"
)

// Options stores configuration settings.
type Options struct {
	dataDir        string // Directory location of local storage
	httpAddress    string // HTTP server address
	raftTCPAddress string // Address for communication between raftNode nodes
	bootstrap      bool   // Start as master or not
	joinAddress    string // Leader raft node address to join
}

// NewOptions initializes an Options struct with values from command-line flags.
func NewOptions() *Options {
	httpAddress := flag.String("http", "127.0.0.1:6000", "HTTP server address")
	raftTCPAddress := flag.String("raft", "127.0.0.1:7000", "raftNode TCP address")
	node := flag.String("node", "node1", "raftNode node name")
	bootstrap := flag.Bool("bootstrap", false, "Start as raftNode cluster master")
	joinAddress := flag.String("join", "", "Join address for raftNode cluster")
	flag.Parse()

	return &Options{
		dataDir:        "./" + *node, // Prefix the node name with `./` assuming it's a directory
		httpAddress:    *httpAddress,
		raftTCPAddress: *raftTCPAddress,
		bootstrap:      *bootstrap,
		joinAddress:    *joinAddress,
	}
}
