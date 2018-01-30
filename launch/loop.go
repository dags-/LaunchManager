package launch

import (
	"bufio"
	"os"
)

func (m *Manager) Run() {
	go m.serve()
	m.PrintCommands()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		go m.ProcessCommand(scanner.Text())
	}
}