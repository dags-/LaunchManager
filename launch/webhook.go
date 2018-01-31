package launch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const webhookUrl = "https://discordapp.com/api/webhooks/"

type Webhook struct {
	Content  string `json:"content"`
	Username string `json:"username"`
	Avatar   string `json:"avatar_url"`
}

func PostWebook(prefs WebhookPrefs, status Status) {
	url := fmt.Sprint(webhookUrl, prefs.Id, "/", prefs.Token)
	wh := Webhook{
		Content:  fmt.Sprint("Status: ", status),
		Username: prefs.Name,
		Avatar:   prefs.Avatar,
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(&wh)
	if err != nil {
		fmt.Println("Webhook Encode Err:", err)
		return
	}

	_, err = http.Post(url, "application/json", &body)
	if err != nil {
		fmt.Println("Webhook Post Err:", err)
	}
}
