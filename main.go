//go:generate rice embed-go
package main

import (
	"github.com/GeertJohan/go.rice"
	"github.com/dags-/LaunchManager/launch"
)

func main() {
	box := rice.MustFindBox("_assets")
	manager := launch.NewManager( box)
	manager.Run()
}
