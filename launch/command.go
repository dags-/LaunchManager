package launch

import (
	"sync"

	"github.com/pkg/errors"
)

type Command func() (error)

type DefaultCommand func(input string) (error)

type Commands struct {
	lock     sync.RWMutex
	notFound error
	fallback DefaultCommand
	commands map[string]Command
}

func NewCommands() (*Commands) {
	notFound := errors.New("command not recognised")
	return &Commands{
		notFound: notFound,
		commands: make(map[string]Command),
		fallback: func(input string) error {
			return notFound
		},
	}
}

func (c *Commands) Call(command string) (error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cmd, ok := c.commands[command]
	if !ok {
		return c.fallback(command)
	}

	// execute and return any errors
	e := cmd()
	return e
}

func (c *Commands) Register(id string, cmd Command) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.commands[id] = cmd
}

func (c *Commands) SetFallback(def DefaultCommand) {
	c.fallback = def
}

func (c *Commands) NotFound() (error) {
	return c.notFound
}
