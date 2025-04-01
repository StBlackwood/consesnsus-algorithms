package main

import (
	"explore_consensus/raft"
	"time"
)

func main() {
	node1 := raft.NewNode("8001", []string{"8002", "8003"})
	node2 := raft.NewNode("8002", []string{"8001", "8003"})
	node3 := raft.NewNode("8003", []string{"8001", "8002"})

	go node1.Start()
	go node2.Start()
	go node3.Start()

	time.Sleep(5 * time.Second) // let election settle

	node1.ReplicateCommand(map[string]string{"x": "1"})
	node2.ReplicateCommand(map[string]string{"x": "2"})
	node3.ReplicateCommand(map[string]string{"x": "3"})
}
