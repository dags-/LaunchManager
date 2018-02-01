package web

import (
	"context"
	"sync"

	"github.com/gorilla/websocket"
)

type Session struct {
	id        string
	auth      bool
	connected bool
	ctx       context.Context
	conn      *websocket.Conn
}

type Sessions struct {
	lock    sync.RWMutex
	backing map[string]*Session
}

func (s *Sessions) ForEach(f func(conn websocket.Conn)) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, s := range s.backing {
		if s != nil && s.conn != nil {
			go f(*s.conn)
		}
	}
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
