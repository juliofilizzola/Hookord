package utils

import "fmt"

const (
	ColorOrange   = 0xe67e22
	ColorGrey     = 0x95a5a6
	ColorGreen    = 0x2ecc71
	ColorBlue     = 0x3498db
	ColorPurple   = 0x9b59b6
	ColorRed      = 0xe74c3c
	ColorYellow   = 0xf1c40f
	ColorDarkGrey = 0x7f8c8d
)

func ColorToHex(colorCode int) string {
	return fmt.Sprintf("#%06x", colorCode)
}

var (
	ColorOrangeText   = ColorToHex(ColorOrange)
	ColorGreyText     = ColorToHex(ColorGrey)
	ColorGreenText    = ColorToHex(ColorGreen)
	ColorBlueText     = ColorToHex(ColorBlue)
	ColorPurpleText   = ColorToHex(ColorPurple)
	ColorRedText      = ColorToHex(ColorRed)
	ColorYellowText   = ColorToHex(ColorYellow)
	ColorDarkGreyText = ColorToHex(ColorDarkGrey)
)
