package github

import (
	"fmt"

	"github.com/juliofilizzola/Hookord/internal/core"
)

func mapPushToEvent(push *PushPayload, evt *core.Event) {
	evt.Repository.Name = push.Repository.Name
	evt.Repository.FullName = push.Repository.FullName
	evt.Repository.URL = push.Repository.HTMLURL
	evt.Repository.Owner = push.Repository.Owner.Login

	evt.Title = fmt.Sprintf("🔥 Push em %s", evt.Repository.FullName)
	evt.URL = push.Compare
	evt.Author.Name = push.Pusher.Name
	evt.Author.Avatar = push.Pusher.AvatarURL

	if len(push.Commits) == 1 {
		evt.Description = fmt.Sprintf("Commits %", push.Commits[0].Msg)
	}

	if len(push.Commits) > 1 {
		evt.Description = fmt.Sprintf("%d commits de %s", len(push.Commits), push.Pusher.Name)
	}
}

func mapPullRequestToEvent(pr *PullRequestPayload, evt *core.Event) {
	evt.Repository.Name = pr.Repo.Name
	evt.Repository.FullName = pr.Repo.FullName
	evt.Repository.URL = pr.Repo.HTMLURL
	evt.Repository.Owner = pr.Repo.Owner.Login

	statusEmoji := mapPRActionEmoji(pr.Action)
	evt.Title = fmt.Sprintf("%s PR #%d: %s", statusEmoji, pr.Number, pr.PullRequest.Title)
	evt.URL = pr.PullRequest.HTMLURL
	evt.Description = fmt.Sprintf("Status: %s | Autor: %s", pr.PullRequest.State, pr.PullRequest.User.Login)
	evt.Author.Name = pr.Sender.Login
	evt.Author.Avatar = pr.Sender.AvatarURL
}

func mapPRActionEmoji(action string) string {
	switch action {
	case "opened":
		return "🆕"
	case "closed":
		return "❌"
	case "merged":
		return "✅"
	case "reopened":
		return "🔄"
	default:
		return "📝"
	}
}
