package raft

func (rn *Node) HandleAppendEntries(req AppendEntriesRequest) AppendEntriesResponse {
	if req.Term < rn.currentTerm {
		return AppendEntriesResponse{Term: rn.currentTerm, Success: false}
	}

	rn.currentTerm = req.Term
	// TODO: Check consistency, append new entries, update commitIndex

	return AppendEntriesResponse{Term: rn.currentTerm, Success: true}
}

func (rn *Node) HandleRequestVote(req RequestVoteRequest) RequestVoteResponse {
	if req.Term < rn.currentTerm {
		return RequestVoteResponse{Term: rn.currentTerm, VoteGranted: false}
	}

	if rn.votedFor == "" || rn.votedFor == req.CandidateID {
		// TODO: Check if candidate’s log is at least as up-to-date
		rn.votedFor = req.CandidateID
		rn.currentTerm = req.Term
		return RequestVoteResponse{Term: req.Term, VoteGranted: true}
	}

	return RequestVoteResponse{Term: rn.currentTerm, VoteGranted: false}
}
