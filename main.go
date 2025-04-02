package main

import (
	"explore_consensus/raft"
	"explore_consensus/zookeeper/weak_hashring"
	"fmt"
	"strconv"
	"time"
)

func main() {
	//RunRaftExperiment()
	RunZKWeakHashringExperiment()
}

func RunRaftExperiment() {
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

func RunZKWeakHashringExperiment() {
	nodeID1 := "8001"
	h := weak_hashring.NewHasher()
	onChange := func(nodes []string) {
		fmt.Println("Current cluster nodes:", nodes)
		h.UpdateNodes(nodes)
	}

	emptyOnChange := func(nodes []string) {
		fmt.Println("Current cluster nodes from dummy nodes:", nodes)
	}

	zkServers := []string{"localhost:2181"}
	_, err := weak_hashring.NewCluster(zkServers, nodeID1, onChange)
	if err != nil {
		panic(err)
	}

	nodeIds := 8002
	for {
		key := "some-key"
		node, err := h.GetNodeForKey(key)
		if err == nil {
			fmt.Printf("Key '%s' handled by node: %s\n", key, node)
		}
		time.Sleep(3 * time.Second)

		_, err = weak_hashring.NewCluster(zkServers, strconv.Itoa(nodeIds), emptyOnChange)
		nodeIds++
		if nodeIds > 8007 {
			return
		}
	}
}
