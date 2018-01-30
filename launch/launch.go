package launch

import (
	"bufio"
	"fmt"
)

func doLaunch(m *Manager) {
	cancelled := false
	// clean-up that should happen after every doLaunch
	defer func() {
		cancelled = true
		m.setProcess(nil)
		m.setStatus(stopped)
		m.setExecutor(emptyExecutor)
		m.PrintCommands()
	}()

	// setup process & input/outputs
	cmd := m.cmd()
	in, err := cmd.StdinPipe()
	if err != nil {
		m.onMessage(fmt.Sprint("Input Err: ", err))
		return
	}

	out, err := cmd.StdoutPipe()
	if err != nil {
		m.onMessage(fmt.Sprint("Output Err: ", err))
		return
	}

	// start process
	if err = cmd.Start(); err != nil {
		m.onMessage(fmt.Sprint("Start Err: ", err))
		return
	}

	// set status as 'starting' and attach current process
	m.setStatus(starting)
	m.setProcess(cmd.Process)

	// start the restart schedule
	go doSchedule(m, &cancelled)

	// read output from the process and pass to onMessage
	go func() {
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			text := scanner.Text()
			m.onMessage(text)
		}
	}()

	// callback that accepts commands and relays to the current process's input
	m.setExecutor(func(s string) (error) {
		_, er := fmt.Fprintln(in, s)
		return er
	})

	// wait for process to finish or crash and handle appropriately
	if err = cmd.Wait(); err != nil {
		m.onCrash(err)
	}
}