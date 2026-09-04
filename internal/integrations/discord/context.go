package discord

const (
	FooterText    = "GitHub ↔ Discord Notification Hookord"
	FooterIconURL = "https://hookord-bp.s3.us-east-1.amazonaws.com/hookord_github.2.png"
)

const (
	FIX   = "fix"
	HOT   = "hot"
	DOC   = "doc"
	CHORE = "chore"
)

func (integration *Integration) Name() string {
	return "discord"
}
