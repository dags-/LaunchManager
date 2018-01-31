//go:generate rice embed-go /i github.com/dags-/LaunchManager/server
package main

import (
	"github.com/dags-/LaunchManager/launch"
)

func main() {
	manager := launch.NewManager()
	manager.Run()
}