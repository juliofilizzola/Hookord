package core

import "time"

type Event struct {
	Id        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`

	Source string `json:"source"`

	Type string `json:"type"`

	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		URL      string `json:"url"`
		Owner    string `json:"owner"`
	} `json:"repository"`

	Payload any `json:"payload,omitempty"`

	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`

	Author struct {
		Name   string `json:"username"`
		Avatar string `json:"avatar_url"`
	} `json:"author"`

	Status string `json:"status"`
}
