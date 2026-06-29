package domain

import "time"

type Repository struct {
	Name     string
	FullName string
	URL      string
	Owner    string
}

type Author struct {
	Name   string
	Avatar string
}

type Event struct {
	Id         string `json:"id"`
	Timestamp  time.Time
	Source     string
	Type       string
	Repository Repository
	Payload    any
	Author     Author
	Status     string
}
