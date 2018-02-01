package web

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

func handleLogin(s *Server) (http.HandlerFunc) {
	return handleErr(func(w http.ResponseWriter, r *http.Request) (error) {
		// hash user address to use as 'state'
		token := getId(r)
		session := &Session{
			id:   token,
			auth: false,
			ctx:  context.Background(),
		}

		s.sessions.Add(session)

		// clean up any un-auth'd sessions after 30 secs
		time.AfterFunc(time.Second*30, func() {
			if !session.auth {
				s.sessions.Del(session.id)
			}
		})

		// redirect to discord for user to authenticate
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

		// check session exists
		se, ok := s.sessions.Get(state)
		if !ok {
			return errors.New("no session exists")
		}

		// check state is valid for the session
		if state != getId(r) {
			s.sessions.Del(state)
			return errors.New("invalid state")
		}

		// exchange auth code for api token
		token, err := s.conf.Exchange(se.ctx, code)
		if err != nil {
			return err
		}

		// get user info for the given token
		client := s.conf.Client(se.ctx, token)
		req, err := client.Get(me)
		if err != nil {
			return err
		}

		user := struct{ ID string `json:"id"` }{}
		err = json.NewDecoder(req.Body).Decode(&user)
		if err != nil {
			return err
		}

		// check user id is in the users.json file
		if !isValidUser(user.ID) {
			s.sessions.Del(se.id)
			return errors.New("not authorised")
		}

		// passed authentication, redirect to console
		se.auth = true
		redirect := fmt.Sprint("/console", "/", se.id)
		http.Redirect(w, r, redirect, 302)
		return nil
	})
}

func getId(r *http.Request) (string) {
	sh := sha256.New()
	sh.Write([]byte(r.RemoteAddr))
	d := sh.Sum(nil)
	i := binary.BigEndian.Uint64(d)
	return strconv.FormatInt(int64(i), 36)
}
