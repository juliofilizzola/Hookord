package notification

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func SendDiscordNotification(webhookURL, message string) error {
	payload := map[string]string{"content": message}
	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post(webhookURL, "application/json", io.NopCloser(bytes.NewReader(jsonPayload)))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}
