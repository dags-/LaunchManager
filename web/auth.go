package web

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type DiscordUser struct {
	ID string `json:"id"`
}

func handleLogin(s *Server) (http.HandlerFunc) {
	return handleErr(func(w http.ResponseWriter, r *http.Request) (error) {
		// hash user address to use as 'state'
		id := getId(r)
		session := s.sessions.Must(id)

		// clean up unused or un-auth'd sessions
		time.AfterFunc(time.Second*60, func() {
			if session.getState() != connected {
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

		// get the session for the given state
		se, err := s.getSession(state, r)
		if err != nil {
			return err
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

		var user DiscordUser
		err = json.NewDecoder(req.Body).Decode(&user)
		if err != nil {
			return err
		}

		// check user id is in the users.json file
		if !isValidUser(user.ID) {
			return errors.New("not authorised")
		}

		// passed authentication, redirect to console
		se.setState(authd)
		redirect := fmt.Sprint("/console", "/", se.id)
		http.Redirect(w, r, redirect, 302)
		return nil
	})
}

func getId(r *http.Request) (string) {
	sh := sha256.New()
	sh.Write([]byte(getIp(r)))
	d := sh.Sum(nil)
	i := binary.BigEndian.Uint64(d)
	return strconv.FormatInt(int64(i), 36)
}

func getIp(r *http.Request) (string) {
	s := r.RemoteAddr
	i := strings.IndexRune(s, ':')
	if i > 0 {
		return s[:i]
	}
	return s
}