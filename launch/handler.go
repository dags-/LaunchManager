package launch

import (
	"fmt"
	"regexp"
	"time"

	"github.com/dags-/LaunchManager/server"
)

var (
	startCheck = regexp.MustCompile(`.*Done \(\d+\.\d+s\)! .*`)
)

// handles status updates
func (m *Manager) onStatus(s Status) {
	fmt.Println(">", s)
	go PostWebook(m.config.Webhook, s)
	go sendMessage(m, "> " + s.String())
}

// handles messages from the running process
func (m *Manager) onMessage(s string) {
	fmt.Println(s)

	go sendMessage(m, s)

	if m.getStatus() == Starting && startCheck.MatchString(s) {
		m.setStatus(Started)
	}
}

// handles process crashes (may be invoked by a call to m.Kill())
func (m *Manager) onCrash(err error) {
	if m.getStatus() != Killed {
		m.setStatus(Crashed)
		time.Sleep(m.getCrashWait())
		go m.Restart()
	}
}

func sendMessage(m *Manager, msg string) {
	if m.server == nil {
		return
	}
	m.server.Outbound <- server.Message{Type: "console", Content: msg}
}