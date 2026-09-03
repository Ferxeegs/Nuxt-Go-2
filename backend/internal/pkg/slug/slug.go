// Package slug generates URL-safe slugs from arbitrary titles.
package slug

import (
	"regexp"
	"strings"
)

var (
	nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
	trimDash = regexp.MustCompile(`^-+|-+$`)
)

func Make(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlnum.ReplaceAllString(s, "-")
	s = trimDash.ReplaceAllString(s, "")
	return s
}
