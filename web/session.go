package web

import (
	"context"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type Session struct {
	id   string
	auth bool
	ctx  context.Context
	conn *websocket.Conn
}

type Sessions struct {
	lock    sync.RWMutex
	backing map[string]*Session
}

func (s *Sessions) ForEach(f func(s *Session)) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, s := range s.backing {
		if s != nil {
			f(s)
		}
	}
}

func (s *Sessions) Get(id string) (*Session, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
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
	fmt.Println("removing session:", id)
	delete(s.backing, id)
}
