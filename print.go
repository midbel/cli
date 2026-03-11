package cli

import (
	"strings"
	"unicode/utf8"
)

func Center(str string, width int) string {
	size := utf8.RuneCountInString(str)
	if size >= width {
		return str
	}
	var (
		padding = width - size
		left = padding / 2
		right = padding - left
	)

	return strings.Repeat(" ", left) + str + strings.Repeat(" ", right)
}

func MarkBool(b bool) string {
	if b {
		return "\u2713"
	}
	return "\u2717"
}