package notification

import "encoding/json"

// Estrutura para receber o payload do GitHub
// Simplificada para PRs

type PullRequestEvent struct {
	Action      string `json:"action"`
	PullRequest struct {
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"pull_request"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

func ParsePullRequestEvent(body []byte) (*PullRequestEvent, error) {
	var event PullRequestEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
