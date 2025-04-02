package raft

import (
	"fmt"
	"net/rpc"
	"time"
)

func (rn *Node) HandleAppendEntries(req AppendEntriesRequest) AppendEntriesResponse {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if req.Term < rn.currentTerm {
		return AppendEntriesResponse{Term: rn.currentTerm, Success: false}
	}

	rn.state = Follower
	rn.currentTerm = req.Term
	rn.lastHeartbeat = time.Now()
	rn.votedFor = ""
	// Basic log consistency check
	if req.PrevLogIndex >= 0 {
		if req.PrevLogIndex >= len(rn.log) ||
			rn.log[req.PrevLogIndex].Term != req.PrevLogTerm {
			return AppendEntriesResponse{Term: rn.currentTerm, Success: false}
		}
	}

	// Append new entries
	for i, entry := range req.Entries {
		index := req.PrevLogIndex + 1 + i
		if index < len(rn.log) {
			rn.log[index] = entry
		} else {
			rn.log = append(rn.log, entry)
		}
	}

	// Commit if possible
	if req.LeaderCommit > rn.commitIndex {
		rn.commitIndex = min(req.LeaderCommit, len(rn.log)-1)
	}

	return AppendEntriesResponse{Term: rn.currentTerm, Success: true}
}

func (rn *Node) HandleRequestVote(req RequestVoteRequest) RequestVoteResponse {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if req.Term < rn.currentTerm {
		return RequestVoteResponse{Term: rn.currentTerm, VoteGranted: false}
	}

	// Step down if new term
	if req.Term > rn.currentTerm {
		rn.currentTerm = req.Term
		rn.votedFor = ""
		rn.state = Follower
	}

	if rn.votedFor != "" && rn.votedFor != req.CandidateID { // if the candidate with most votes dies down, new election term will start
		return RequestVoteResponse{Term: rn.currentTerm, VoteGranted: false}
	}

	lastTerm := rn.lastLogTerm()
	lastIndex := len(rn.log) - 1

	upToDate := req.LastLogTerm > lastTerm ||
		(req.LastLogTerm == lastTerm && req.LastLogIndex >= lastIndex)
	if !upToDate {
		return RequestVoteResponse{Term: rn.currentTerm, VoteGranted: false}
	}

	rn.votedFor = req.CandidateID
	return RequestVoteResponse{Term: req.Term, VoteGranted: true}

}

// RPC handlers (must be exported and use pointer receiver)

// AppendEntries RPC handler
func (rn *Node) AppendEntries(req AppendEntriesRequest, res *AppendEntriesResponse) error {
	*res = rn.HandleAppendEntries(req)
	return nil
}

// RequestVote RPC handler
func (rn *Node) RequestVote(req RequestVoteRequest, res *RequestVoteResponse) error {
	*res = rn.HandleRequestVote(req)
	return nil
}

// Client-side RPC calls
func (rn *Node) sendAppendEntries(peer string, req AppendEntriesRequest) (AppendEntriesResponse, error) {
	var res AppendEntriesResponse
	client, err := rpc.Dial("tcp", ":"+peer)
	if err != nil {
		return res, fmt.Errorf("dial error: %w", err)
	}
	defer client.Close()

	err = client.Call(peer+".AppendEntries", req, &res)
	return res, err
}

func (rn *Node) SendRequestVote(peer string, req RequestVoteRequest) (RequestVoteResponse, error) {
	var res RequestVoteResponse
	client, err := rpc.Dial("tcp", ":"+peer)
	if err != nil {
		return res, fmt.Errorf("dial error: %w", err)
	}
	defer client.Close()

	err = client.Call(peer+".RequestVote", req, &res)
	return res, err
}
