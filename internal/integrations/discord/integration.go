package discord

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/juliofiliizzola/hookord/internal/domain"
	"github.com/juliofiliizzola/hookord/internal/integrations"
)

var (
	// ErrClientUnavailable is returned when the Discord client is nil or disconnected.
	ErrClientUnavailable = errors.New("discord: client is unavailable")

	// ErrChannelNotConfigured is returned when no channel is configured for the event category.
	ErrChannelNotConfigured = errors.New("discord: channel not configured")

	// ErrNilEvent is returned when the event payload or entity is nil.
	ErrNilEvent = errors.New("discord: event payload is nil")
)

// Integration implements integrations.Integration for Discord.
type Integration struct {
	client DiscordClient
	repo   domain.MessageRepository
	cfg    Config
}

// New creates a new Discord Integration instance.
func New(cfg Config, repo domain.MessageRepository, client DiscordClient) *Integration {
	return &Integration{
		client: client,
		repo:   repo,
		cfg:    cfg,
	}
}

// NewWithToken creates a new Discord Integration by initializing a live DiscordClient with the provided token.
func NewWithToken(cfg Config, repo domain.MessageRepository) (*Integration, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client, err := NewSessionClient(cfg.Token)
	if err != nil {
		return nil, err
	}

	return New(cfg, repo, client), nil
}

// Name returns the identifier of this integration.
func (i *Integration) Name() string {
	return "discord"
}

// HandlePullRequest processes a GitHub Pull Request event and sends or updates the Discord message.
func (i *Integration) HandlePullRequest(ctx context.Context, event *integrations.PullRequestEvent) error {
	if event == nil || event.PullRequest == nil {
		return ErrNilEvent
	}

	if i.client == nil {
		return ErrClientUnavailable
	}

	channelID := i.cfg.GetChannel("pull_requests")
	if channelID == "" {
		return fmt.Errorf("%w: pull_requests", ErrChannelNotConfigured)
	}

	entityID := strconv.FormatInt(event.PullRequest.GetID(), 10)

	mapping, err := i.repo.GetMapping(ctx, entityID)
	if err != nil {
		return err
	}

	if mapping != nil {
		if event.TotalReviews == 0 && mapping.TotalReviews > 0 {
			event.TotalReviews = mapping.TotalReviews
		}
		if event.TotalReviewers == 0 && mapping.TotalReviewers > 0 {
			event.TotalReviewers = mapping.TotalReviewers
		}
	}

	content := BuildPullRequestContent(event.PullRequest)
	embed := BuildPullRequestEmbed(event)

	if mapping == nil || mapping.DiscordMessageID == "" {
		data := &discordgo.MessageSend{
			Content: content,
			Embed:   embed,
		}

		msg, err := i.client.ChannelMessageSendComplex(channelID, data)
		if err != nil {
			return err
		}

		repoFullName := ""
		if event.Repository != nil {
			repoFullName = event.Repository.GetFullName()
		}

		if mapping == nil {
			mapping = &domain.MessageMapping{
				EntityID:   entityID,
				Repository: repoFullName,
				Reviewers:  make(map[string]bool),
			}
		}

		mapping.DiscordMessageID = msg.ID
		mapping.DiscordChannelID = channelID
		mapping.LastStatus = event.PullRequest.GetState()
		mapping.TotalReviews = event.TotalReviews
		mapping.TotalReviewers = event.TotalReviewers

		return i.repo.SaveMapping(ctx, mapping)
	}

	targetChannel := mapping.DiscordChannelID
	if targetChannel == "" {
		targetChannel = channelID
	}

	if _, err := i.client.ChannelMessageEditEmbed(targetChannel, mapping.DiscordMessageID, embed); err != nil {
		return err
	}

	mapping.LastStatus = event.PullRequest.GetState()
	return i.repo.SaveMapping(ctx, mapping)
}

// HandleIssue processes a GitHub Issue event and sends or updates the Discord message.
func (i *Integration) HandleIssue(ctx context.Context, event *integrations.IssueEvent) error {
	if event == nil || event.Issue == nil {
		return ErrNilEvent
	}

	if i.client == nil {
		return ErrClientUnavailable
	}

	channelID := i.cfg.GetChannel("issues")
	if channelID == "" {
		return fmt.Errorf("%w: issues", ErrChannelNotConfigured)
	}

	entityID := strconv.FormatInt(event.Issue.GetID(), 10)

	mapping, err := i.repo.GetMapping(ctx, entityID)
	if err != nil {
		return err
	}

	content := BuildIssueContent(event.Issue)
	embed := BuildIssueEmbed(event)

	if mapping == nil || mapping.DiscordMessageID == "" {
		data := &discordgo.MessageSend{
			Content: content,
			Embed:   embed,
		}

		msg, err := i.client.ChannelMessageSendComplex(channelID, data)
		if err != nil {
			return err
		}

		repoFullName := ""
		if event.Repository != nil {
			repoFullName = event.Repository.GetFullName()
		}

		if mapping == nil {
			mapping = &domain.MessageMapping{
				EntityID:   entityID,
				Repository: repoFullName,
			}
		}

		mapping.DiscordMessageID = msg.ID
		mapping.DiscordChannelID = channelID
		mapping.LastStatus = event.Issue.GetState()

		return i.repo.SaveMapping(ctx, mapping)
	}

	targetChannel := mapping.DiscordChannelID
	if targetChannel == "" {
		targetChannel = channelID
	}

	if _, err := i.client.ChannelMessageEditEmbed(targetChannel, mapping.DiscordMessageID, embed); err != nil {
		return err
	}

	mapping.LastStatus = event.Issue.GetState()
	return i.repo.SaveMapping(ctx, mapping)
}

// Close gracefully closes the Discord client connection.
func (i *Integration) Close() error {
	if i.client != nil {
		return i.client.Close()
	}
	return nil
}
