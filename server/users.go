package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func isValidUser(id string) (bool) {
	data, err := ioutil.ReadFile("users.json")
	if err != nil {
		fmt.Println(err)
		return false
	}

	var users map[string]bool
	err = json.Unmarshal(data, &users)
	if err != nil {
		fmt.Println(err)
		return false
	}

	_, ok := users[id]
	return ok
}
