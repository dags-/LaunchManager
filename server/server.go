package server

import (
	"fmt"
	"net/http"
	"sync"

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

func NewServer(auth *oauth2.Config) (*Server) {
	box := rice.MustFindBox("../_assets")

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

func (s *Server) Start(port int) {
	s.port = port
	m := mux.NewRouter()
	m.HandleFunc("/auth", handleAuth(s))
	m.HandleFunc("/login", handleLogin(s))
	m.HandleFunc("/feed/{id}", s.handleFeed())
	m.HandleFunc("/console/{id}", s.handleConsole())
	go http.ListenAndServe(fmt.Sprintf(":%v", s.port), m)
}

func handleErr(f func(w http.ResponseWriter, r *http.Request) error) (http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			http.Error(w, err.Error(), 401)
		}
	}
}
