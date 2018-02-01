package launch

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/dags-/LaunchManager/web"
	"github.com/pkg/errors"
)

// the application state
type Manager struct {
	lock     sync.RWMutex
	status   Status
	config   Config
	commands *Commands
	server   *web.Server
	input    io.WriteCloser
	process  *os.Process
}

func NewManager() (*Manager) {
	m := &Manager{status: Stopped}
	m.config = loadConfig()
	m.commands = NewCommands()
	m.commands.SetFallback(m.Fallback)
	m.commands.Register("start", m.Start)
	m.commands.Register("stop", m.Stop)
	m.commands.Register("restart", m.Restart)
	m.commands.Register("kill", m.Kill)
	m.commands.Register("reload", m.Reload)
	m.commands.Register("exit", m.Exit)
	m.commands.Register("status", m.Status)
	return m
}

func (m *Manager) getRestartWait() (time.Duration) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	t := m.config.Schedule.Restart
	return time.Duration(t) * time.Minute
}

func (m *Manager) getStatus() (Status) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	s := m.status
	return s
}

func (m *Manager) getCrashWait() (time.Duration) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	t := m.config.Schedule.CrashWait
	return time.Duration(t) * time.Second
}

func (m *Manager) getCommand() (*exec.Cmd) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	var args []string
	args = append(args, "-jar")
	args = append(args, m.config.Launch.Target)
	args = append(args, m.config.Launch.Args...)
	return exec.Command(m.config.Launch.Runtime, args...)
}

func (m *Manager) getWebhook() (string, string, string, string) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	prefs := m.config.Webhook
	return prefs.Id, prefs.Token, prefs.Name, prefs.Avatar
}

func (m *Manager) hasProcess() (bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.process != nil
}

func (m *Manager) exec(cmd string) (error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.input != nil {
		_, err := fmt.Fprintln(m.input, cmd)
		return err
	}
	return errors.New("no process currently attached")
}

func (m *Manager) reloadConfig() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.config = loadConfig()
}

func (m *Manager) setProcess(p *os.Process, w io.WriteCloser) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.process = p
	m.input = w
}

func (m *Manager) setStatus(s Status) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.status = s
	m.onStatus(s)
}
