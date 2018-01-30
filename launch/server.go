package launch

import (
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	box     *rice.Box
	lock    sync.RWMutex
	upgrade websocket.Upgrader
	clients map[string]Client
}

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type Client struct {
	id   string
	name string
	conn *websocket.Conn
}

func (m *Manager) serve() {
	m.Lock()
	port := m.config.Server.Port
	m.Unlock()
	mx := http.NewServeMux()
	mx.HandleFunc("/feed", m.server.feed)
	mx.HandleFunc("/console", m.server.webui)
	go m.handleActions()
	panic(http.ListenAndServe(fmt.Sprintf(":%v", port), mx))
}

func (s *Server) feed(w http.ResponseWriter, r *http.Request) {
	if con, err := s.upgrade.Upgrade(w, r, nil); err == nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		client := Client{id: "todo", name: "todo", conn: con}
		client.conn.SetCloseHandler(func(code int, text string) error {
			s.lock.Lock()
			defer s.lock.Unlock()
			delete(s.clients, client.id)
			return client.conn.Close()
		})
		s.clients[client.id] = client
	} else {
		fmt.Println(err)
	}
}

func (s *Server) webui(w http.ResponseWriter, r *http.Request) {
	data := s.box.MustBytes("index.html")
	w.Write(data)
}

func (s *Server) handleMessage(text string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	msg := &Message{Type: "console", Content: text}
	for _, c := range s.clients {
		c.conn.WriteJSON(msg)
	}
}

func (m *Manager) handleActions() {
	s := m.server
	for {
		s.lock.Lock()
		for _, c := range s.clients {
			var msg Message
			err := c.conn.ReadJSON(&msg)
			if err == nil {
				go m.ProcessCommand(msg.Content)
			}
		}
		s.lock.Unlock()
		time.Sleep(time.Second)
	}
}
