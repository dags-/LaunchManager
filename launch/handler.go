package launch

import (
	"fmt"
	"regexp"
	"time"
)

var (
	startCheck = regexp.MustCompile(`.*Done \(\d+\.\d+s\)! .*`)
)

func (m *Manager) onStatus(s string) {
	fmt.Println(">", s)
	go PostStatus(m.config.Webhook, s)
	go m.server.handleMessage(fmt.Sprintln(">", s))
}

func (m *Manager) onMessage(s string) {
	fmt.Println(s)

	go m.server.handleMessage(s)

	if m.getStatus() == starting && startCheck.MatchString(s) {
		m.setStatus(started)
	}
}

func (m *Manager) onCrash(err error) {
	if m.getStatus() != killed {
		m.setStatus(crashed)
		time.Sleep(m.getCrashWait())
		go m.Restart()
	}
}