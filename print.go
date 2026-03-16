package cli

import (
	"strings"
	"unicode/utf8"
)

var (
	checkMarker = "\u2713"
	crossMarker = "\u2717"
)

func Center(str string, width int) string {
	size := utf8.RuneCountInString(str)
	if size >= width {
		return str
	}
	var (
		padding = width - size
		left    = padding / 2
		right   = padding - left
	)

	return strings.Repeat(" ", left) + str + strings.Repeat(" ", right)
}

func Right(str string, width int) string {
	size := utf8.RuneCountInString(str)
	if size >= width {
		return str
	}
	padding := width - size
	return strings.Repeat(" ", padding) + str
}

func MarkBool(b bool) string {
	if b {
		return checkMarker
	}
	return crossMarker
}
