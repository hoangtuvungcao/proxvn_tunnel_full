package tunnel

import (
	"sync"
)

type ControlMessageQueue struct {
	mu       sync.Mutex
	messages []Message
}

func NewControlMessageQueue() *ControlMessageQueue {
	return &ControlMessageQueue{
		messages: make([]Message, 0),
	}
}

func (q *ControlMessageQueue) Push(msg Message) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.messages = append(q.messages, msg)
}

func (q *ControlMessageQueue) PopAll() []Message {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.messages) == 0 {
		return nil
	}
	res := q.messages
	q.messages = make([]Message, 0)
	return res
}

func (q *ControlMessageQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.messages)
}
