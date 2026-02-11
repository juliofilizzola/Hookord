package github

import "github.com/rs/zerolog"

type PushPayload struct {
	Ref        string     `json:"ref"`
	Before     string     `json:"before"`
	After      string     `json:"after"`
	Repository Repository `json:"repository"`
	Pusher     Pusher     `json:"pusher"`
	Commits    []Commit   `json:"commits"`
	Compare    string     `json:"compare"`
}

type PullRequestPayload struct {
	Action      string      `json:"action"`
	Number      int         `json:"number"`
	Repo        Repository  `json:"repository"`
	PullRequest PullRequest `json:"pull_request"`
	Sender      User        `json:"sender"`
}

type Repository struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
	Owner    Owner  `json:"owner"`
}

type Owner struct {
	Login string `json:"login"`
}

type User struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

type Pusher struct {
	Name string `json:"name"`
}

type Commit struct {
	ID  string `json:"id"`
	Msg string `json:"message"`
}

type PullRequest struct {
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
	State   string `json:"state"`
	User    User   `json:"user"`
}

type Provider struct {
	logger zerolog.Logger
}
