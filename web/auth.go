package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

func handleLogin(s *Server) (http.HandlerFunc) {
	return handleErr(func(w http.ResponseWriter, r *http.Request) (error) {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}

		session := &Session{
			id: id.String(),
			auth: false,
			ctx: context.Background(),
		}

		s.sessions.Add(session)

		time.AfterFunc(time.Second * 30, func() {
			if !session.auth {
				s.sessions.Del(session.id)
			}
		})

		url := s.conf.AuthCodeURL(session.id, oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, 302)
		return nil
	})
}

func handleAuth(s *Server) (http.HandlerFunc) {
	const me = "https://discordapp.com/api/users/@me"
	return handleErr(func(w http.ResponseWriter, r *http.Request) (error) {
		state := r.FormValue("state")
		code := r.FormValue("code")

		se, ok := s.sessions.Get(state)
		if !ok {
			return errors.New("no session exists")
		}

		token, err := s.conf.Exchange(se.ctx, code)
		if err != nil {
			return err
		}

		client := s.conf.Client(se.ctx, token)
		req, err := client.Get(me)
		if err != nil {
			return err
		}

		user := struct {ID string `json:"id"`}{}
		err = json.NewDecoder(req.Body).Decode(&user)
		if err != nil {
			return err
		}

		if !isValidUser(user.ID) {
			s.sessions.Del(se.id)
			return errors.New("not authorised")
		}

		se.auth = true
		redirect := fmt.Sprint("/console", "/", se.id)
		http.Redirect(w, r, redirect, 301)

		return nil
	})
}