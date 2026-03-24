package cli

import (
	"strconv"
	"strings"
)

func Success(str string) string {
	return Fg(Green).Apply(str)
}

func Error(str string) string {
	return Fg(Red).Brightened().Bolded().Apply(str)
}

func Warning(str string) string {
	return Fg(Yellow).Bolded().Apply(str)
}

func Info(str string) string {
	return Fg(Blue).Apply(str)
}

type Color int

const (
	noCol Color = iota
	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

type Style struct {
	Fg     Color
	Bg     Color
	Bright bool
	Bold   bool
	Dim    bool
}

func Fg(color Color) Style {
	var s Style
	s.Fg = color
	return s
}

func (s Style) Apply(str string) string {
	if s.isZero() {
		return str
	}
	var b strings.Builder

	b.WriteString("\033[")

	writeCode := func(code, part int) {
		if part > 0 {
			b.WriteByte(';')
		}
		b.WriteString(strconv.Itoa(code))
	}

	var part int
	if s.Bold {
		writeCode(boldCode, part)
		part++
	}
	if s.Dim {
		writeCode(dimCode, part)
		part++
	}

	if s.Fg != noCol {
		writeCode(s.getForegroundBase()+int(s.Fg), part)
		part++
	}

	if s.Bg != noCol {
		writeCode(s.getBackgroundBase()+int(s.Bg), part)
	}

	b.WriteByte('m')
	b.WriteString(str)
	b.WriteString("\033[0m")

	return b.String()
}

func (s Style) Default() Style {
	s.Bold = false
	s.Dim = false
	s.Bright = false
	return s
}

func (s Style) Dimmed() Style {
	s.Dim = true
	return s
}

func (s Style) Bolded() Style {
	s.Bold = true
	return s
}

func (s Style) Brightened() Style {
	s.Bright = true
	return s
}

func (s Style) isZero() bool {
	return s.Fg == noCol && s.Bg == noCol && !s.Bright && !s.Bold && !s.Dim
}

func (s Style) getBackgroundBase() int {
	if s.Bright {
		return bgBaseBright
	}
	return bgBase
}

func (s Style) getForegroundBase() int {
	if s.Bright {
		return fgBaseBright
	}
	return fgBase
}

const (
	fgBase       = 30
	fgBaseBright = 90
	bgBase       = 40
	bgBaseBright = 100
	boldCode     = 1
	dimCode      = 2
)
