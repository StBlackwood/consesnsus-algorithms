package raft

import (
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type Node struct {
	id          string
	Peers       []string
	currentTerm int
	votedFor    string
	log         []LogEntry
	state       string
	commitIndex int
	lastApplied int

	lastHeartbeat time.Time
	mu            sync.Mutex
}

func NewNode(id string, peers []string) *Node {
	return &Node{
		id:    id,
		Peers: peers,
		state: Follower,
	}
}

const (
	Follower  = "follower"
	Candidate = "candidate"
	Leader    = "leader"
)

func (rn *Node) Start() error {
	err := rpc.RegisterName(rn.id, rn)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "localhost:"+rn.id)
	if err != nil {
		return err
	}
	go rpc.Accept(listener)

	go rn.electionTimer()
	rn.lastHeartbeat = time.Now()

	log.Printf("[%s] started as %s", rn.id, rn.state)
	return nil
}

func (rn *Node) electionTimer() {
	for {
		time.Sleep(time.Duration(50+rand.Intn(50)) * time.Millisecond)

		rn.mu.Lock()
		state := rn.state
		elapsed := time.Since(rn.lastHeartbeat)
		rn.mu.Unlock()

		// If still follower and no heartbeat, trigger election
		if state == Follower && elapsed > 300*time.Millisecond {
			log.Printf("[%s] election timeout - becoming candidate", rn.id)
			rn.startElection()
		}
	}
}

func (rn *Node) startElection() {
	rn.mu.Lock()
	rn.state = Candidate
	rn.currentTerm++
	rn.votedFor = rn.id
	term := rn.currentTerm
	rn.mu.Unlock()

	votes := 1
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, peer := range rn.Peers {
		wg.Add(1)
		go func(peer string) {
			defer wg.Done()
			req := RequestVoteRequest{
				Term:         term,
				CandidateID:  rn.id,
				LastLogIndex: len(rn.log) - 1,
				LastLogTerm:  rn.lastLogTerm(),
			}
			res, err := rn.SendRequestVote(peer, req)
			if err == nil && res.VoteGranted {
				mu.Lock()
				votes++
				mu.Unlock()
			} else {
				if err != nil {
					log.Printf("error in election %v", err)
				} else {
					log.Printf("[%s] vote not granted for id: %s", peer, rn.id)
				}
			}
		}(peer)
	}
	wg.Wait()

	if votes > len(rn.Peers)/2 {
		log.Printf("[%s] wins election with %d votes", rn.id, votes)
		rn.mu.Lock()
		rn.state = Leader
		rn.votedFor = ""
		rn.mu.Unlock()
		go rn.heartbeatLoop()
	}
}

func (rn *Node) lastLogTerm() int {
	if len(rn.log) == 0 {
		return 0
	}
	return rn.log[len(rn.log)-1].Term
}

func (rn *Node) heartbeatLoop() {
	for rn.state == Leader {
		for _, peer := range rn.Peers {
			go func(peer string) {
				rn.mu.Lock()
				req := AppendEntriesRequest{
					Term:         rn.currentTerm,
					LeaderID:     rn.id,
					PrevLogIndex: len(rn.log) - 1,
					PrevLogTerm:  rn.lastLogTerm(),
					Entries:      []LogEntry{}, // heartbeat (empty)
					LeaderCommit: rn.commitIndex,
				}
				rn.mu.Unlock()

				_, err := rn.sendAppendEntries(peer, req)
				if err != nil {
					log.Printf("[%s] heartbeat to %s failed: %v", rn.id, peer, err)
				}
			}(peer)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (rn *Node) ReplicateCommand(cmd interface{}) {
	rn.mu.Lock()
	if rn.state != Leader {
		log.Printf("[%s] not the leader, ignoring command", rn.id)
		rn.mu.Unlock()
		return
	}

	entry := LogEntry{Term: rn.currentTerm, Command: cmd}
	rn.log = append(rn.log, entry)
	rn.mu.Unlock()

	// Attempt to send to all peers
	for _, peer := range rn.Peers {
		go func(peer string) {
			rn.mu.Lock()
			req := AppendEntriesRequest{
				Term:         rn.currentTerm,
				LeaderID:     rn.id,
				PrevLogIndex: len(rn.log) - 2,
				PrevLogTerm:  rn.lastLogTerm(),
				Entries:      []LogEntry{entry},
				LeaderCommit: rn.commitIndex,
			}
			rn.mu.Unlock()

			res, err := rn.sendAppendEntries(peer, req)
			if err == nil && res.Success {
				log.Printf("[%s] replicated command to %s", rn.id, peer)
			}
		}(peer)
	}
}
