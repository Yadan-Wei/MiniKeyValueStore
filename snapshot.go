package main

import "github.com/hashicorp/raft"

// impletation of hashicorp/raftNode
type Snapshot struct {
	cm *KV
}

// Persist saves the FSM Snapshot out to the given sink.
func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	// convert local storage to bytes
	snapshotBytes, err := s.cm.Marshal()
	if err != nil {
		_ = sink.Cancel()
		return err
	}

	if _, err := sink.Write(snapshotBytes); err != nil {
		_ = sink.Cancel()
		return err
	}

	if err := sink.Close(); err != nil {
		_ = sink.Cancel()
		return err
	}
	return nil
}

func (s *Snapshot) Release() {}
