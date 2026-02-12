package discord

import (
	"fmt"
	"strconv"
	"time"

	"github.com/juliofilizzola/Hookord/internal/core"
	"github.com/juliofilizzola/Hookord/internal/providers"
	"github.com/juliofilizzola/Hookord/internal/providers/github"
)

func BuildEmbed(event core.Event) *Embed {

	embed := &Embed{
		Title:       event.Title,
		Description: event.Description,
		URL:         event.URL,
		Timestamp:   event.Timestamp.Format(time.RFC3339),
		Color:       (&github.Provider{}).EventColor(providers.GithubEventType(event.Type)),
		Footer:      nil,
		Auth:        nil,
		Fields:      nil,
	}

	embed.Footer = &EmbedFooter{
		Text: fmt.Sprintf("%s/%s", event.Repository.Owner, event.Repository.Name),
	}

	embed.Auth = &EmbedAuthor{
		Name:    event.Author.Name,
		IconURL: event.Author.Avatar,
		URL:     fmt.Sprintf("https://github.com/%s", event.Author.Name),
	}

	switch event.Type {
	case "push":
		push, ok := event.Payload.(github.PushPayload)
		if !ok {
			break
		}

		embed.Fields = append(embed.Fields, EmbedField{
			Name:   "Commits",
			Value:  strconv.Itoa(len(push.Commits)),
			Inline: true,
		})

	case "pull_request":
		pr, ok := event.Payload.(github.PullRequestPayload)
		if !ok {
			break
		}

		embed.Fields = []EmbedField{
			{Name: "Status", Value: pr.PullRequest.State, Inline: true},
			{Name: "#", Value: strconv.Itoa(pr.Number), Inline: true},
			{Name: "Ação", Value: pr.Action, Inline: true},
		}
	}

	return embed
}
