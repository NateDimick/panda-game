package global

import "strings"

var AvailableStyles []Tailwinder = []Tailwinder{
	GreenBBTheme,
	YellowBBTheme,
	PinkBBTheme,
	PlainTheme,
	ButtonBorder,
	ButtonSpacing,
	Centered,
	BodyContainer,
	FlexContainer,
}

// this is so tailwind has something to parse for generated classes, like bg-<color>
func DumpAllStyles() string {
	var sb strings.Builder
	for _, s := range AvailableStyles {
		sb.WriteString(s.Classes())
		sb.WriteRune(' ')
	}
	return sb.String()
}
