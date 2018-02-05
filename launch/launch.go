package launch

import (
	"bufio"
	"fmt"
)

// handles the launch process
func launch(m *Manager) {
	// clean-up that should happen after every launch
	defer func() {
		m.setProcess(nil, nil)
		m.setStatus(Stopped)
		m.PrintCommands()
	}()

	// setup process & input/outputs
	cmd := m.getCommand()
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

	// set status as 'starting' and attach current process/writer
	m.setStatus(Starting)
	m.setProcess(cmd.Process, in)

	// start the restart schedule
	go scheduleRestart(m)

	// read output from the process and pass to onMessage
	go func() {
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			text := scanner.Text()
			m.onMessage(text)
		}
	}()

	// wait for process to finish, handle crash appropriately
	if err = cmd.Wait(); err != nil {
		m.onCrash(err)
	}
}
