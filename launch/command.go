package launch

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
)

// processes commandline input
func (m *Manager) ProcessCommand(cmd string) (error) {
	switch cmd {
	case "kill":
		m.Kill()
		break
	case "exit":
		m.Exit()
		break
	case "start":
		m.Start()
		break
	case "restart":
		m.Restart()
		break
	case "stop":
		m.Stop()
		break
	case "reload":
		m.Reload()
		break
	case "status":
		m.Status()
		break
	default:
		if m.getStatus() == Started {
			return m.Exec(cmd)
		}
		m.onMessage(fmt.Sprint("Command: ", cmd, " not recognised!"))
	}
	return nil
}

// launches a new process
func (m *Manager) Start() {
	if m.getStatus() == Stopped {
		doLaunch(m)
	} else {
		m.onMessage("Already running. Please wait until process has stopped")
	}
}

// stops an active process
func (m *Manager) Stop() {
	if m.getStatus() == Started {
		m.setStatus(Stopping)
		m.Say("Now stopping")
		m.Exec("save-all")
		m.Exec("stop")
	} else {
		m.onMessage("Process has not yet started")
	}
}

// stops the current process if active, then starts a new one
func (m *Manager) Restart() {
	status := m.getStatus()
	if status != Started && status != Stopped {
		m.onMessage("Cannot restart whilst in the '" + status.String() + "' state")
		return
	}

	// wait for starting phase to complete
	for m.getStatus() == Starting {
		time.Sleep(time.Second)
	}

	// do a stop if process is running
	if m.getStatus() == Started {
		m.Stop()
	}

	// wait for process to have stopped
	for m.getStatus() != Stopped {
		time.Sleep(time.Second)
	}

	// start a new process
	m.Start()
}

// kills an active process without doing the normal (and safer) stop process
func (m *Manager) Kill() {
	m.Lock()
	defer m.Unlock()
	if m.process != nil {
		m.status = Killed
		m.onStatus(Killed)
		m.process.Kill()
		m.process.Release()
		m.process = nil
	}
}

// reload the config
func (m *Manager) Reload() {
	m.Lock()
	defer m.Unlock()
	m.config = loadConfig()
}

// exit the application, performs a stop or kills the active process if times out
func (m *Manager) Exit() {
	if m.getStatus() != Stopped && m.getStatus() != Crashed {
		m.Stop()
	}

	start := time.Now()
	timeout := time.Duration(5) * time.Second
	for m.getStatus() != Stopped {
		if time.Since(start) > timeout {
			m.Kill()
			break
		}
		time.Sleep(time.Second)
	}

	os.Exit(0)
}

// prints the current status
func (m *Manager) Status() {
	m.onMessage("Status: " + m.getStatus().String())
}

// performs a 'say' command
func (m *Manager) Say(message string, args ...interface{}) (error) {
	if len(args) == 0 {
		return m.Exec(fmt.Sprint("say ", message))
	}
	return m.Exec(fmt.Sprint("say ", fmt.Sprintf(message, args...)))
}

// writes a command to the current process' input writer
func (m *Manager) Exec(cmd string) (error) {
	m.RLock()
	defer m.RUnlock()
	if m.input != nil {
		_, err := fmt.Fprintln(m.input, cmd)
		return err
	}
	return errors.New("no process currently attached")
}

// prints the available commands
func (m *Manager) PrintCommands() {
	buf := bytes.Buffer{}
	fmt.Fprintln(&buf, "=============== Launch Manager ==============")
	fmt.Fprintln(&buf, "start | stop | restart | kill | status | exit")
	fmt.Println(buf.String())
}
