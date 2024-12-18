package global

import (
	"fmt"
	"strings"
)

type Tailwinder interface {
	Classes() string
}

func CombineClasses(classes ...Tailwinder) string {
	classStrings := make([]string, 0)
	for _, t := range classes {
		classStrings = append(classStrings, t.Classes())
	}
	return strings.Join(classStrings, " ")
}

type DrawWidth int

const (
	NoWidth DrawWidth = 0
	Pencil  DrawWidth = 1
	Skinny  DrawWidth = 2
	Think   DrawWidth = 4
	Chunky  DrawWidth = 8
)

type Rounding string

const (
	Square Rounding = "rounded-none"
	Subtle Rounding = "rounded"
	Gentle Rounding = "rounded-md"
	Smooth Rounding = "rounded-lg"
	Circle Rounding = "rounded-full"
)

type SideMode int

const (
	Same SideMode = iota
	XY
	UDLR // up, down, left, right
	None
)

type ContainerWidth string

const (
	HalfWidth   ContainerWidth = "w-1/2"
	ThirdWidth  ContainerWidth = "w-1/3"
	FourthWidth ContainerWidth = "w-1/4"
	FullWidth   ContainerWidth = "w-full"
	ScreenWidth ContainerWidth = "w-screen"
)

type ContainerHeight string

const (
	HalfHeight   ContainerHeight = "h-1/2"
	ThirdHeight  ContainerHeight = "h-1/3"
	FourthHeight ContainerHeight = "h-1/4"
	FullHeight   ContainerHeight = "h-full"
	ScreenHeight ContainerHeight = "h-screen"
)

const (
	GreenPrimary    string = "bbgreen"
	GreenSecondary  string = "bbgreenoff"
	YellowPrimary   string = "bbyellow"
	YellowSecondary string = "bbyellowoff"
	PinkPrimary     string = "bbpink"
	PinkSecondary   string = "bbpinkoff"
	Black           string = "black"
	White           string = "white"
	Neutral         string = "neutral"
)

type ColorStyle struct {
	BackgroundColor      string
	BorderColor          string
	TextColor            string
	HoverBackgroundColor string
	HoverTextColor       string
}

func (t *ColorStyle) Classes() string {
	return fmt.Sprintf("bg-%s border-%s text-%s hover:bg-%s hover:text-%s", t.BackgroundColor, t.BorderColor, t.TextColor, t.HoverBackgroundColor, t.HoverTextColor)
}

var PlainTheme = &ColorStyle{
	BackgroundColor:      White,
	BorderColor:          Black,
	TextColor:            Black,
	HoverBackgroundColor: Neutral,
	HoverTextColor:       Black,
}

var GreenBBTheme = &ColorStyle{
	BackgroundColor:      GreenPrimary,
	BorderColor:          GreenSecondary,
	TextColor:            Black,
	HoverBackgroundColor: GreenSecondary,
	HoverTextColor:       Black,
}

var YellowBBTheme = &ColorStyle{
	BackgroundColor:      YellowPrimary,
	BorderColor:          YellowSecondary,
	TextColor:            Black,
	HoverBackgroundColor: YellowSecondary,
	HoverTextColor:       Black,
}

var PinkBBTheme = &ColorStyle{
	BackgroundColor:      PinkPrimary,
	BorderColor:          PinkSecondary,
	TextColor:            Black,
	HoverBackgroundColor: PinkSecondary,
	HoverTextColor:       Black,
}

type BorderStyle struct {
	Width  DrawWidth
	Radius Rounding
}

func (b *BorderStyle) Classes() string {
	w := fmt.Sprintf("border-%d", b.Width)
	if b.Width == Pencil {
		w = "border"
	}
	return fmt.Sprintf("%s %s", w, b.Radius)
}

var ButtonBorder = &BorderStyle{
	Width:  Skinny,
	Radius: Gentle,
}

type SpacingStyle struct {
	MarginAuto  bool
	MarginMode  SideMode
	Margin      []int // -1 for auto
	PaddingMode SideMode
	Padding     []int
}

func (s *SpacingStyle) Classes() string {
	xy := []string{"x-", "y-"}
	udlr := []string{"t-", "b-", "l-", "r-"}
	m := make([]string, 0)
	p := make([]string, 0)
	if s.MarginAuto {
		m = append(m, "m-auto")
	} else if s.MarginMode == XY {
		for i := range 2 {
			m = append(m, fmt.Sprintf("%s%d", xy[i], s.Margin[i]))
			if s.Margin[i] == -1 {
				m[i] = xy[i] + "auto"
			}
		}
	} else if s.MarginMode == UDLR {
		for i := range 4 {
			m = append(m, fmt.Sprintf("%s%d", udlr[i], s.Margin[i]))
			if s.Margin[i] == -1 {
				m[i] = udlr[i] + "auto"
			}
		}
	} else if s.MarginMode == Same {
		m = append(m, fmt.Sprintf("m-%d", s.Margin[0]))
	}

	if s.PaddingMode == XY {
		for i := range 2 {
			p = append(p, fmt.Sprintf("%s%d", xy[i], s.Padding[i]))
		}
	} else if s.PaddingMode == UDLR {
		for i := range 4 {
			p = append(p, fmt.Sprintf("%s%d", udlr[i], s.Padding[i]))
		}
	} else if s.PaddingMode == Same {
		p = append(p, fmt.Sprintf("p-%d", s.Padding[0]))
	}
	return fmt.Sprintf("%s %s", strings.Join(m, " "), strings.Join(p, " "))
}

var Centered = &SpacingStyle{
	MarginAuto:  true,
	PaddingMode: None,
}

var ButtonSpacing = &SpacingStyle{
	PaddingMode: Same,
	Padding:     []int{2},
	MarginAuto:  true,
}

type ContainerStyle struct {
	Width    ContainerWidth
	Height   ContainerHeight
	Position string
	Flex     bool // flex justify-stretch flex-wrap
}

func (c *ContainerStyle) Classes() string {
	f := ""
	if c.Flex {
		f = "flex justify-stretch flex-wrap"
	}
	return fmt.Sprintf("%s %s %s %s", f, c.Width, c.Height, c.Position)
}

var BodyContainer = &ContainerStyle{
	Width:    ScreenWidth,
	Height:   ScreenHeight,
	Position: "static",
}

var FlexContainer = &ContainerStyle{
	Width:    FullWidth,
	Flex:     true,
	Position: "static",
}
