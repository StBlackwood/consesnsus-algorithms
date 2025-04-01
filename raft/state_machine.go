package raft

type StateMachine struct {
	state map[string]string
}

func NewStateMachine() *StateMachine {
	return &StateMachine{state: make(map[string]string)}
}

func (sm *StateMachine) Apply(command interface{}) {
	// This is just a placeholder. In real Raft, the command would be structured.
	if kv, ok := command.(map[string]string); ok {
		for k, v := range kv {
			sm.state[k] = v
		}
	}
}
