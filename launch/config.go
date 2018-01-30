package launch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type WebhookPrefs struct {
	Id     string `json:"id"`
	Token  string `json:"token"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type JavaPrefs struct {
	Runtime string   `json:"runtime"`
	Target  string   `json:"target"`
	Args    []string `json:"args"`
}

type SchedulePrefs struct {
	Restart   int `json:"restart"`
	CrashWait int `json:"crash"`
}

type ServerPrefs struct {
	Port int `json:"port"`
}

type Config struct {
	Webhook  WebhookPrefs  `json:"webhook"`
	Launch   JavaPrefs     `json:"launch"`
	Schedule SchedulePrefs `json:"schedule"`
	Server   ServerPrefs   `json:"server"`
}

func loadConfig() (Config) {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		config := defaultConfig()
		writeConfig(&config)
		return config
	} else {
		var config Config
		if err := json.Unmarshal(data, &config); err != nil {
			fmt.Println(err)
		}
		return config
	}
}

func writeConfig(c *Config) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile("config.json", data, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
}

func defaultConfig() (Config) {
	return Config{
		Schedule: SchedulePrefs{
			Restart:   1,
			CrashWait: 10,
		},
		Launch: JavaPrefs{
			Runtime: "java",
			Target:  "my_jar.jar",
			Args:    []string{"-Xmx2G", "-Xms1G"},
		},
		Webhook: WebhookPrefs{
			Name:   "ServerStatus",
			Token:  "",
			Avatar: "",
		},
		Server: ServerPrefs{
			Port: 8123,
		},
	}
}
