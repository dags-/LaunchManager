//go:generate rice embed-go
//go:generate go build
package main

import (
	"github.com/GeertJohan/go.rice"
	"github.com/dags-/LaunchManager/launch"
)

func main() {
	box := rice.MustFindBox("_assets")
	manager := launch.NewManager(box)
	manager.Run()
}
