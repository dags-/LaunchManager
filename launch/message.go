package launch

import (
	"fmt"
	"time"

	"github.com/dags-/LaunchManager/web"
)

// handles status updates
func (m *Manager) onStatus(s Status) {
	fmt.Println(">", s)
	go sendWebhook(m, s.String())
	go sendMessage(m, "> "+s.String())
}

// handles messages from the running process
func (m *Manager) onMessage(s string) {
	fmt.Println(s)

	go sendMessage(m, s)

	if m.getStatus() == Starting && startCheck.MatchString(s) {
		m.setStatus(Started)
	}
}

// handles error messages
func (m *Manager) onError(e error) {
	if e != nil {
		m.onMessage(e.Error())
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
	m.server.Outbound <- web.Message{Type: "console", Content: msg}
}

func sendWebhook(m *Manager, msg string) {
	id, token, name, avatar := m.getWebhook()
	web.PostWebhook(id, token, &web.Webhook{
		Content:  msg,
		Username: name,
		Avatar:   avatar,
	})
}
