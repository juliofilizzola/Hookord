package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/juliofilizzola/Hookord/internal/core"
	"github.com/rs/zerolog"
)

func NewClient(webhookURL string, logger *zerolog.Logger) *Client {
	return &Client{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger: logger.
			With().
			Str("provider", "discord").
			Str("webhook_url", webhookURL[:30]+"...").
			Logger(),
	}
}

func (c *Client) SendMessage(event core.Event) error {
	embed := BuildEmbed(event)
	if embed == nil {
		return fmt.Errorf("failed to build embed")
	}
	payload := Webhook{
		Embeds:   []Embed{*embed},
		Username: fmt.Sprintf("Hookord - %s", event.Repository.Name),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewBuffer(body))

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.logger.Error().Msgf("failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != 204 {
		c.logger.Error().Msgf("Discord webhook returned status code %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Name() string {
	return "discord"
}
