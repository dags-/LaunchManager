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

type Token struct {
	Port int
	ID   string
}

// http connections for console page
func (s *Server) handleConsole() (http.HandlerFunc) {
	text := s.box.MustString("index.html")
	temp := template.Must(template.New("console").Parse(text))
	return handleErr(func(w http.ResponseWriter, r *http.Request) (error) {
		vars := mux.Vars(r)
		sid := vars["id"]

		se, err := s.getSession(sid, r)
		if err != nil {
			s.sessions.Del(sid)
			return err
		}

		if se.getState() != authd {
			s.sessions.Del(sid)
			return errors.New("session is not authenticated")
		}

		se.setState(connected)
		return temp.Execute(w, &Token{Port: s.port, ID: se.id})
	})
}

// websocket connections for console feed
func (s *Server) handleFeed() (http.HandlerFunc) {
	upgrader := websocket.Upgrader{}
	return handleErr(func(w http.ResponseWriter, r *http.Request) (error) {
		vars := mux.Vars(r)
		sid := vars["id"]

		se, err := s.getSession(sid, r)
		if err != nil {
			s.sessions.Del(sid)
			return err
		}

		if se.getState() != connected {
			s.sessions.Del(sid)
			return errors.New("session is not authenticated")
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			s.sessions.Del(sid)
			return err
		}

		// allow return visit within 20 secs, otherwise remove the session
		conn.SetCloseHandler(func(code int, text string) (error) {
			time.AfterFunc(time.Second*20, func() {
				if se, ok := s.sessions.Get(sid); ok && se.getState() != connected {
					s.sessions.Del(sid)
				}
			})
			return errors.New(text)
		})

		go read(s.Inbound, conn)
		go write(s.Outbound, conn)

		return nil
	})
}

func read(c chan Message, conn *websocket.Conn) {
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			return
		}
		c <- msg
	}
}

func write(c chan Message, conn *websocket.Conn) {
	for {
		msg := <-c
		err := conn.WriteJSON(&msg)
		if err != nil {
			return
		}
	}
}
