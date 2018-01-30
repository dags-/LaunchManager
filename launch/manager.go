package launch

import (
	"fmt"
	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/websocket"
	"os"
	"os/exec"
	"sync"
	"time"
)

const (
	starting = "starting"
	started  = "started"
	stopping = "stopping"
	stopped  = "stopped"
	crashed  = "crashed"
	killed   = "killed"
)

type Manager struct {
	lock      sync.RWMutex
	status    string
	process   *os.Process
	server    Server
	config    Config
	onExecute func(string) (error)
}

func NewManager(b *rice.Box) (*Manager) {
	return &Manager{
		config:    loadConfig(),
		status:    stopped,
		onExecute: emptyExecutor,
		server: Server{
			box:     b,
			upgrade: websocket.Upgrader{},
			clients: make(map[string]Client),
		},
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

func (m *Manager) setProcess(p *os.Process) {
	m.Lock()
	defer m.Unlock()
	m.process = p
}

func (m *Manager) getStatus() (string) {
	m.RLock()
	defer m.RUnlock()
	s := m.status
	return s
}

func (m *Manager) setStatus(s string) {
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

func (m *Manager) cmd() (*exec.Cmd) {
	m.RLock()
	defer m.RUnlock()
	var args []string
	args = append(args, "-jar")
	args = append(args, m.config.Launch.Target)
	args = append(args, m.config.Launch.Args...)
	return exec.Command(m.config.Launch.Runtime, args...)
}

func (m *Manager) setExecutor(e func(string) (error)) {
	m.Lock()
	defer m.Unlock()
	m.onExecute = e
}

func emptyExecutor(s string) (error) {
	fmt.Println("empty:", s)
	return nil
}
