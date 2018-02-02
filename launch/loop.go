package launch

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/dags-/LaunchManager/web"
	"golang.org/x/oauth2"
)

// starts application loop
func (m *Manager) Run() {
	m.PrintCommands()

	m.server = startServer(m.config, m.box)
	go processInbound(m)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		go processCommand(m, scanner.Text())
	}
}

// creates the server, starts it in a new goroutine, and returns it's pointer
func startServer(config Config, box *rice.Box) (*web.Server) {
	s := web.NewServer(&oauth2.Config{
		ClientID:     config.Server.ClientId,
		ClientSecret: config.Server.ClientSecret,
		RedirectURL:  fmt.Sprintf(config.Server.RedirectUri, config.Server.Port),
		Endpoint:     web.DiscordEndpoints(),
		Scopes:       []string{"identify"},
	}, box)
	s.Start(config.Server.Port)
	return s
}

// processes incoming messages from the server websockets, rate limited at 250ms
func processInbound(m *Manager) {
	ticker := time.NewTicker(time.Millisecond * 250)
	defer ticker.Stop()
	for range ticker.C {
		msg := <-m.server.Inbound
		if msg.Type == "command" {
			go processCommand(m, msg.Content)
		}
	}
}

// processes command input, printing any error messages thrown
func processCommand(m *Manager, cmd string) {
	err := m.commands.Call(cmd)
	if err != nil {
		m.onError(err)
	}
}
