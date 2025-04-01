package raft

import (
	"log"
)

type Node struct {
	id          string
	peers       []string
	currentTerm int
	votedFor    string
	log         []LogEntry
	state       string
	commitIndex int
	lastApplied int
}

func NewRaftNode(id string, peers []string) *Node {
	return &Node{
		id:    id,
		peers: peers,
		state: "follower",
	}
}

func (rn *Node) Start() error {
	log.Printf("Raft node %s started as follower", rn.id)
	// TODO: Start RPC servers, timers for elections, etc.
	return nil
}
