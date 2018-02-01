package web

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

func PostWebhook(id, token string, webhook *Webhook) {
	url := fmt.Sprint(webhookUrl, id, "/", token)

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(webhook)
	if err != nil {
		fmt.Println("Webhook Encode Err:", err)
		return
	}

	_, err = http.Post(url, "application/json", &body)
	if err != nil {
		fmt.Println("Webhook Post Err:", err)
	}
}
