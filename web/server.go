package web

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

type Server struct {
	lock     sync.RWMutex
	box      *rice.Box
	conf     *oauth2.Config
	port     int
	sessions Sessions
	Inbound  chan Message
	Outbound chan Message
}

func NewServer(auth *oauth2.Config, box *rice.Box) (*Server) {
	return &Server{
		box:      box,
		conf:     auth,
		Inbound:  make(chan Message),
		Outbound: make(chan Message),
		sessions: Sessions{
			backing: make(map[string]*Session),
		},
	}
}

func DiscordEndpoints() (oauth2.Endpoint) {
	return oauth2.Endpoint{
		AuthURL:  "https://discordapp.com/api/oauth2/authorize",
		TokenURL: "https://discordapp.com/api/oauth2/token",
	}
}

func (s *Server) Start(port int) {
	s.port = port
	m := mux.NewRouter()
	m.HandleFunc("/auth", handleAuth(s)).Methods("GET")
	m.HandleFunc("/login", handleLogin(s)).Methods("GET")
	m.HandleFunc("/feed/{id}", s.handleFeed()).Methods("GET")
	m.HandleFunc("/console/{id}", s.handleConsole()).Methods("GET")
	srv := &http.Server{
		Handler:        m,
		Addr:           fmt.Sprintf(":%v", s.port),
		WriteTimeout:   5 * time.Second,
		ReadTimeout:    5 * time.Second,
		MaxHeaderBytes: 2096,
	}
	go srv.ListenAndServe()
}

func (s *Server) getSession(sid string, r *http.Request) (*Session, error) {
	se, ok := s.sessions.Get(sid)
	if !ok {
		return nil, errors.New("no active session")
	}

	id := getId(r)
	if id != se.id {
		return nil, errors.New("invalid id provided")
	}

	return se, nil
}

func handleErr(f func(w http.ResponseWriter, r *http.Request) error) (http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			http.Error(w, err.Error(), 401)
		}
	}
}
