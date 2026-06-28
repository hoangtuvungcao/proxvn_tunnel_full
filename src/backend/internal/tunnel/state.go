package tunnel

import (
	"fmt"
	"sync"
)

type ConnState string

const (
	StateInit           ConnState = "INIT"
	StateConnecting     ConnState = "CONNECTING"
	StateTLSHandshake   ConnState = "TLS_HANDSHAKE"
	StateAuthenticating ConnState = "AUTHENTICATING"
	StateRegistering    ConnState = "REGISTERING"
	StateActive         ConnState = "ACTIVE"
	StateDegraded       ConnState = "DEGRADED"
	StateReconnecting   ConnState = "RECONNECTING"
	StateClosed         ConnState = "CLOSED"
)

type StateMachine struct {
	mu      sync.RWMutex
	current ConnState
}

func NewStateMachine() *StateMachine {
	return &StateMachine{current: StateInit}
}

func (sm *StateMachine) Get() ConnState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

func (sm *StateMachine) TransitionTo(next ConnState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !isValidTransition(sm.current, next) {
		return fmt.Errorf("invalid state transition from %s to %s", sm.current, next)
	}
	sm.current = next
	return nil
}

func isValidTransition(from, to ConnState) bool {
	if from == to {
		return true
	}
	if from == StateClosed {
		return false // Closed is a terminal state
	}

	switch from {
	case StateInit:
		return to == StateConnecting || to == StateClosed
	case StateConnecting:
		return to == StateTLSHandshake || to == StateReconnecting || to == StateClosed
	case StateTLSHandshake:
		return to == StateAuthenticating || to == StateReconnecting || to == StateClosed
	case StateAuthenticating:
		return to == StateRegistering || to == StateReconnecting || to == StateClosed
	case StateRegistering:
		return to == StateActive || to == StateReconnecting || to == StateClosed
	case StateActive:
		return to == StateDegraded || to == StateReconnecting || to == StateClosed
	case StateDegraded:
		return to == StateActive || to == StateReconnecting || to == StateClosed
	case StateReconnecting:
		return to == StateConnecting || to == StateClosed
	}
	return false
}
