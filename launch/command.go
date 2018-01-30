package launch

import (
	"fmt"
	"os"
	"time"
)

func (m *Manager) Start() {
	if m.getStatus() == stopped {
		doLaunch(m)
	} else {
		fmt.Println("Already running. Please wait until process has stopped")
	}
}

func (m *Manager) Stop() {
	if m.getStatus() == started {
		m.setStatus(stopping)
		m.Say("Stopping")
		m.Exec("save-all")
		m.Exec("stop")
	} else {
		fmt.Println("Process has not yet started")
	}
}

func (m *Manager) Restart() {
	for m.getStatus() == starting {
		time.Sleep(time.Second)
	}

	if m.getStatus() == started {
		m.Stop()
	}

	for m.getStatus() != stopped {
		time.Sleep(time.Second)
	}

	m.Start()
}

func (m *Manager) Kill() {
	m.Lock()
	defer m.Unlock()
	if m.process != nil {
		m.status = killed
		m.onStatus(killed)
		m.process.Kill()
		m.process.Release()
		m.process = nil
	}
}

func (m *Manager) Reload() {
	m.Lock()
	defer m.Unlock()
	m.config = loadConfig()
}

func (m *Manager) Exit() {
	if m.getStatus() != stopped && m.getStatus() != crashed {
		m.Stop()
	}

	start := time.Now()
	timeout := time.Duration(5) * time.Second
	for m.getStatus() != stopped {
		if time.Since(start) > timeout {
			m.Kill()
			break
		}
		time.Sleep(time.Second)
	}

	os.Exit(0)
}

func (m *Manager) Status() {
	fmt.Println("Status:", m.getStatus())
}

func (m *Manager) Say(message string, args ...interface{}) (error) {
	if len(args) == 0 {
		return m.Exec(fmt.Sprint("say ", message))
	}
	return m.Exec(fmt.Sprint("say ", fmt.Sprintf(message, args...)))
}

func (m *Manager) Exec(cmd string) (error) {
	m.RLock()
	defer m.RUnlock()
	return m.onExecute(cmd)
}

func (m *Manager) PrintCommands() {
	fmt.Println("=============== Launch Manager ==============")
	fmt.Println("start | stop | restart | kill | status | exit")
}

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
		if m.getStatus() == started {
			return m.Exec(cmd)
		}
		fmt.Println("Command:", cmd, "not recognised!")
	}
	return nil
}
