package core

type OutputPort interface {
	SendMessage(event Event) error
	Name() string
}
