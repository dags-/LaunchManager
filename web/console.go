package web

import (
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type Auth struct {
	Port  int
	Token string
}

// http connections for console page
func (s *Server) handleConsole() (http.HandlerFunc) {
	text := s.box.MustString("index.html")
	temp := template.Must(template.New("console").Parse(text))
	return handleErr(func(w http.ResponseWriter, r *http.Request) (error) {
		vars := mux.Vars(r)
		state := vars["id"]

		if getId(r) != state {
			return errors.New("state rejected")
		}

		se, ok := s.sessions.Get(state)
		if !ok || se == nil || !se.auth {
			return errors.New("not authorised")
		}

		return temp.Execute(w, &Auth{Port: s.port, Token: state})
	})
}

// websocket connections for console feed
func (s *Server) handleFeed() (http.HandlerFunc) {
	go handleInbound(s)
	go handleOutbound(s)

	upgrader := websocket.Upgrader{}
	return handleErr(func(w http.ResponseWriter, r *http.Request) (error) {
		vars := mux.Vars(r)
		state := vars["id"]

		if getId(r) != state {
			return errors.New("state rejected")
		}

		se, ok := s.sessions.Get(state)
		if !ok || se == nil || !se.auth {
			return errors.New("not authorised")
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			s.sessions.Del(state)
			return err
		}

		se.conn = conn
		se.connected = true

		conn.SetCloseHandler(func(code int, text string) (error) {
			time.AfterFunc(time.Second * 20, func() {
				if se, ok := s.sessions.Get(state); ok && !se.connected {
					s.sessions.Del(state)
				}
			})
			return errors.New(text)
		})

		return nil
	})
}

// reads messages from client sessions and relays to the inbound channel
func handleInbound(s *Server) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		// read messages into buffer
		// ForEach holds a lock on sessions so we don't want to do anything blocking inside!
		s.sessions.ForEach(func(conn websocket.Conn) {
			var msg Message
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			s.Inbound <- msg
		})
	}
}

// relays outbound messages to client sessions
func handleOutbound(s *Server) {
	for {
		msg := <-s.Outbound
		s.sessions.ForEach(func(conn websocket.Conn) {
			conn.WriteJSON(&msg)
		})
	}
}
