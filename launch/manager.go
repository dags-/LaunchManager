package launch

import (
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/dags-/LaunchManager/server"
)

// the application state
type Manager struct {
	lock    sync.RWMutex
	status  Status
	config  Config
	server  *server.Server
	input   io.WriteCloser
	process *os.Process
}

func NewManager() (*Manager) {
	config := loadConfig()
	return &Manager{
		status: Stopped,
		config: config,
	}
}

func (m *Manager) Lock() {
	m.lock.Lock()
}

func (m *Manager) Unlock() {
	m.lock.Unlock()
}

func (m *Manager) RLock() {
	m.lock.RLock()
}

func (m *Manager) RUnlock() {
	m.lock.RUnlock()
}

func (m *Manager) setProcess(p *os.Process, w io.WriteCloser) {
	m.Lock()
	defer m.Unlock()
	m.process = p
	m.input = w
}

func (m *Manager) getStatus() (Status) {
	m.RLock()
	defer m.RUnlock()
	s := m.status
	return s
}

func (m *Manager) setStatus(s Status) {
	m.Lock()
	defer m.Unlock()
	m.status = s
	m.onStatus(s)
}

func (m *Manager) getRestartWait() (time.Duration) {
	m.RLock()
	defer m.RUnlock()
	t := m.config.Schedule.Restart
	return time.Duration(t) * time.Minute
}

func (m *Manager) getCrashWait() (time.Duration) {
	m.RLock()
	defer m.RUnlock()
	t := m.config.Schedule.CrashWait
	return time.Duration(t) * time.Second
}

func (m *Manager) getCommand() (*exec.Cmd) {
	m.RLock()
	defer m.RUnlock()
	var args []string
	args = append(args, "-jar")
	args = append(args, m.config.Launch.Target)
	args = append(args, m.config.Launch.Args...)
	return exec.Command(m.config.Launch.Runtime, args...)
}