package discord

const (
	FIX   = "fix"
	HOT   = "hot"
	DOC   = "doc"
	CHORE = "chore"
)

func (integration *Integration) Name() string {
	return "discord"
}
