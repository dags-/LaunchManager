package launch

import (
	"bufio"
	"fmt"
	"os"

	"github.com/dags-/LaunchManager/server"
	"golang.org/x/oauth2"
)

// starts application loop
func (m *Manager) Run() {
	m.PrintCommands()

	m.server = startServer(m.config)
	go processInbound(m)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		go m.ProcessCommand(scanner.Text())
	}
}

func startServer(config Config) (*server.Server) {
	auth := &oauth2.Config{
		ClientID:     config.Server.ClientId,
		ClientSecret: config.Server.ClientSecret,
		RedirectURL: fmt.Sprintf(config.Server.RedirectUri, config.Server.Port),
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discordapp.com/api/oauth2/authorize",
			TokenURL: "https://discordapp.com/api/oauth2/token",
		},
		Scopes: []string{"identify"},
	}
	s := server.NewServer(auth)
	s.Start(config.Server.Port)
	return s
}

func processInbound(m *Manager) {
	for {
		msg := <- m.server.Inbound
		if msg.Type == "command" {
			go m.ProcessCommand(msg.Content)
		}
	}
}