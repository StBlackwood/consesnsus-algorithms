package main

import (
	"explore_consensus/raft"
	"time"
)

func main() {
	go raft.NewNode("8001", []string{"8002", "8003"}).Start()
	go raft.NewNode("8002", []string{"8001", "8003"}).Start()
	go raft.NewNode("8003", []string{"8001", "8002"}).Start()

	time.Sleep(5 * time.Second) // let election settle

	node := raft.NewNode("8001", []string{"8002", "8003"})
	node.ReplicateCommand(map[string]string{"x": "42"})
}
