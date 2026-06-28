package tunnel

import (
	"sync"
	"testing"
)

func TestNewStateMachineStartsAtInit(t *testing.T) {
	sm := NewStateMachine()
	if got := sm.Get(); got != StateInit {
		t.Fatalf("new state machine = %q, want %q", got, StateInit)
	}
}

func TestValidHappyPath(t *testing.T) {
	sm := NewStateMachine()
	path := []ConnState{
		StateConnecting,
		StateTLSHandshake,
		StateAuthenticating,
		StateRegistering,
		StateActive,
	}
	for _, next := range path {
		if err := sm.TransitionTo(next); err != nil {
			t.Fatalf("TransitionTo(%q) returned error: %v", next, err)
		}
		if got := sm.Get(); got != next {
			t.Fatalf("after transition, state = %q, want %q", got, next)
		}
	}
}

func TestInvalidTransitionIsRejected(t *testing.T) {
	sm := NewStateMachine()
	// INIT -> ACTIVE is not allowed (must go through the handshake states).
	if err := sm.TransitionTo(StateActive); err == nil {
		t.Fatal("expected error for INIT -> ACTIVE, got nil")
	}
	// State must be unchanged after a rejected transition.
	if got := sm.Get(); got != StateInit {
		t.Fatalf("state changed after rejected transition: %q", got)
	}
}

func TestClosedIsTerminal(t *testing.T) {
	sm := NewStateMachine()
	if err := sm.TransitionTo(StateClosed); err != nil {
		t.Fatalf("INIT -> CLOSED should be allowed: %v", err)
	}
	// Any transition out of CLOSED (other than CLOSED itself) must fail.
	for _, next := range []ConnState{StateConnecting, StateActive, StateReconnecting} {
		if err := sm.TransitionTo(next); err == nil {
			t.Fatalf("CLOSED -> %q should be rejected, got nil error", next)
		}
	}
	// CLOSED -> CLOSED is idempotent and allowed.
	if err := sm.TransitionTo(StateClosed); err != nil {
		t.Fatalf("CLOSED -> CLOSED should be idempotent: %v", err)
	}
}

func TestSameStateTransitionAllowed(t *testing.T) {
	sm := NewStateMachine()
	_ = sm.TransitionTo(StateConnecting)
	if err := sm.TransitionTo(StateConnecting); err != nil {
		t.Fatalf("same-state transition should be allowed: %v", err)
	}
}

func TestDegradedRecoversToActive(t *testing.T) {
	sm := NewStateMachine()
	for _, s := range []ConnState{StateConnecting, StateTLSHandshake, StateAuthenticating, StateRegistering, StateActive, StateDegraded} {
		if err := sm.TransitionTo(s); err != nil {
			t.Fatalf("setup transition to %q failed: %v", s, err)
		}
	}
	if err := sm.TransitionTo(StateActive); err != nil {
		t.Fatalf("DEGRADED -> ACTIVE should be allowed: %v", err)
	}
}

func TestReconnectCycle(t *testing.T) {
	sm := NewStateMachine()
	_ = sm.TransitionTo(StateConnecting)
	if err := sm.TransitionTo(StateReconnecting); err != nil {
		t.Fatalf("CONNECTING -> RECONNECTING should be allowed: %v", err)
	}
	if err := sm.TransitionTo(StateConnecting); err != nil {
		t.Fatalf("RECONNECTING -> CONNECTING should be allowed: %v", err)
	}
}

// TestConcurrentTransitions exercises the mutex under -race.
func TestConcurrentTransitions(t *testing.T) {
	sm := NewStateMachine()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sm.TransitionTo(StateConnecting)
			_ = sm.Get()
			_ = sm.TransitionTo(StateClosed)
		}()
	}
	wg.Wait()
	// After the dust settles the machine must be in a defined state.
	if got := sm.Get(); got != StateConnecting && got != StateClosed {
		t.Fatalf("unexpected final state: %q", got)
	}
}
