package launch

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

)

var (
	startCheck = regexp.MustCompile(`.*Done \(\d+\.\d+s\)! .*`)
)

func (m* Manager) Fallback(command string) (error) {
	if m.getStatus() == Started {
		return m.Exec(command)
	}
	return m.commands.NotFound()
}

// launches a new process
func (m *Manager) Start() (error) {
	if m.getStatus() == Stopped {
		go launch(m)
		return nil
	} else {
		return errors.New("already running. please wait until process has stopped")
	}
}

// stops an active process
func (m *Manager) Stop() (error) {
	if m.getStatus() == Started {
		m.setStatus(Stopping)
		m.Say("Now stopping")
		m.Exec("save-all")
		m.Exec("stop")
		return nil
	} else {
		return errors.New("process has not started yet")
	}
}

// stops the current process if active, then starts a new one
func (m *Manager) Restart() (error) {
	status := m.getStatus()
	if status != Started && status != Stopped {
		return errors.New("cannot restart whilst in the '" + status.String() + "' state")
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
	return m.Start()
}

// kills an active process without doing the normal (and safer) stop process
func (m *Manager) Kill() (error) {
	m.Lock()
	defer m.Unlock()
	if m.process != nil {
		m.status = Killed
		m.onStatus(Killed)
		m.process.Kill()
		m.process.Release()
		m.process = nil
		return nil
	}
	return errors.New("no process attached")
}

// reload the config
func (m *Manager) Reload() (error) {
	m.Lock()
	defer m.Unlock()
	m.config = loadConfig()
	return nil
}

// exit the application, performs a stop or kills the active process if times out
func (m *Manager) Exit() (error) {
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
	return nil
}

// prints the current status
func (m *Manager) Status() (error) {
	m.onMessage("Status: " + m.getStatus().String())
	return nil
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
	fmt.Fprintln(&buf, "=============== Launch Commands ==============")
	fmt.Fprintln(&buf, "start | stop | restart | kill | status | exit")
	fmt.Println(buf.String())
}