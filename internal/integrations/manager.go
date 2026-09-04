package integrations

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
)

// Manager manages and dispatches events to registered integrations.
type Manager struct {
	integrations []Integration
}

// NewManager creates a new Manager instance with the given integrations.
func NewManager(integrations ...Integration) *Manager {
	return &Manager{
		integrations: integrations,
	}
}

// Register adds an integration to the manager.
func (m *Manager) Register(i Integration) {
	m.integrations = append(m.integrations, i)
}

// Integrations returns all currently registered integrations.
func (m *Manager) Integrations() []Integration {
	return m.integrations
}

// HandlePullRequest dispatches a Pull Request event to all registered integrations.
func (m *Manager) HandlePullRequest(ctx context.Context, event *PullRequestEvent) error {
	var errs []error
	for _, integration := range m.integrations {
		if err := integration.HandlePullRequest(ctx, event); err != nil {
			log.Error().
				Err(err).
				Str("integration", integration.Name()).
				Msg("failed to handle pull request event")
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// HandleIssue dispatches an Issue event to all registered integrations.
func (m *Manager) HandleIssue(ctx context.Context, event *IssueEvent) error {
	var errs []error
	for _, integration := range m.integrations {
		if err := integration.HandleIssue(ctx, event); err != nil {
			log.Error().
				Err(err).
				Str("integration", integration.Name()).
				Msg("failed to handle issue event")
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
