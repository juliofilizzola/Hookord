package github

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/juliofilizzola/Hookord/internal/core"
	"github.com/juliofilizzola/Hookord/internal/providers"
)

func (p *Provider) Parse(eventType string, payload []byte) (core.Event, error) {
	evt := core.Event{
		Id:        uuid.NewString(),
		Timestamp: time.Now(),
		Source:    "github",
		Type:      eventType,
	}

	switch providers.GithubEventType(eventType) {
	case providers.GithubPushEvent:
		var push PushPayload

		if err := json.Unmarshal(payload, &push); err != nil {
			return core.Event{}, err
		}

		mapPushToEvent(&push, &evt)
	case providers.GithubPullRequestEvent:
		var pr PullRequestPayload
		if err := json.Unmarshal(payload, &pr); err != nil {
			return core.Event{}, fmt.Errorf("failed to unmarshal pull request payload: %w", err)
		}
		mapPullRequestToEvent(&pr, &evt)
	case providers.GithubPingEvent:
		evt.Title = "GitHub webhook ping recebido"
		evt.Description = "A configuração do webhook está correta e o Hookord está pronto para receber eventos."
		evt.URL = "https://docs.github.com/webhooks"

		return evt, nil
	default:
		p.logger.Warn().Msgf("event type %s not supported", eventType)
		return evt, fmt.Errorf("event type %s not supported", eventType)

	}

	return evt, nil
}

func (p *Provider) EventTypeSupported() []providers.GithubEventType {
	return []providers.GithubEventType{
		providers.GithubPushEvent,
		providers.GithubIssueEvent,
		providers.GithubPullRequestEvent,
		providers.GithubReleaseEvent,
		providers.GithubWorkflowEvent,
		providers.GithubPingEvent,
		providers.GithubDeploymentEvent,
	}
}
