package web

import (
	"context"
	"sync"
	"sync/atomic"
)

const (
	unauthd   int64 = 0
	authd     int64 = 1
	connected int64 = 2
)

type Session struct {
	id    string
	state int64
	ctx   context.Context
}

type Sessions struct {
	lock    sync.RWMutex
	backing map[string]*Session
}

func (s *Session) setState(i int64) {
	atomic.StoreInt64(&s.state, i)
}

func (s *Session) getState() (int64) {
	return atomic.LoadInt64(&s.state)
}

func (s *Sessions) Must(id string) (*Session) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if se, ok := s.backing[id]; ok {
		return se
	}

	se := &Session{
		id:    id,
		state: unauthd,
		ctx:   context.Background(),
	}
	s.backing[id] = se

	return se
}

func (s *Sessions) Get(id string) (*Session, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	se, ok := s.backing[id]
	return se, ok
}

func (s *Sessions) Add(se *Session) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.backing[se.id] = se
}

func (s *Sessions) Del(id string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.backing, id)
}
