package tui

import (
	"regexp"
	"strings"
)

var (
	// {*}bold{*}
	reJiraBold = regexp.MustCompile(`\{\*\}(.*?)\{\*\}`)
	// {_}italic{_}
	reJiraItalic = regexp.MustCompile(`\{_\}(.*?)\{_\}`)
	// {{monospace}}
	reJiraMono = regexp.MustCompile(`\{\{(.*?)\}\}`)
	// {color:#hex}text{color}
	reJiraColor = regexp.MustCompile(`\{color[^}]*\}(.*?)\{color\}`)
	// {quote}text{quote}
	reJiraQuote = regexp.MustCompile(`\{quote\}(.*?)\{quote\}`)
	// {noformat}text{noformat}
	reJiraNoformat = regexp.MustCompile(`\{noformat\}`)
	// {code}...{code} or {code:lang}...{code}
	reJiraCode = regexp.MustCompile(`\{code(?::[^}]*)?\}`)
	// !image.png|options! or !image.png!
	reJiraImage = regexp.MustCompile(`!([^|!\n]+?)(?:\|[^!\n]*)?!`)
	// [text|url]
	reJiraLinkText = regexp.MustCompile(`\[([^|\]]+)\|([^\]]+)\]`)
	// [url]
	reJiraLinkBare = regexp.MustCompile(`\[([^\]|]+)\]`)
	// *bold* (only when surrounded by whitespace or line boundaries)
	reStarBold = regexp.MustCompile(`(?:^|\s)\*(\S.*?\S|\S)\*(?:\s|$)`)
)

func formatJiraMarkup(text string) string {
	// {*}bold{*} → BOLD (terminal bold)
	text = reJiraBold.ReplaceAllString(text, "\033[1m$1\033[22m")

	// {_}italic{_} → italic (terminal dim as approximation)
	text = reJiraItalic.ReplaceAllString(text, "$1")

	// {{monospace}} → `code`
	text = reJiraMono.ReplaceAllString(text, "$1")

	// {color}text{color} → just text
	text = reJiraColor.ReplaceAllString(text, "$1")

	// {quote}text{quote} → text
	text = reJiraQuote.ReplaceAllString(text, "$1")

	// {noformat} → remove marker
	text = reJiraNoformat.ReplaceAllString(text, "")

	// {code} / {code:lang} → remove marker
	text = reJiraCode.ReplaceAllString(text, "")

	// !image.png|opts! → [image: filename]
	text = reJiraImage.ReplaceAllString(text, "[image: $1]")

	// [text|url] → text (url)
	text = reJiraLinkText.ReplaceAllString(text, "$1 ($2)")

	// [url] → url (but preserve [image: ...])
	text = reJiraLinkBare.ReplaceAllStringFunc(text, func(s string) string {
		inner := s[1 : len(s)-1]
		if strings.HasPrefix(inner, "image: ") {
			return s
		}
		return inner
	})

	// *bold* → bold (careful: only match word-bounded)
	text = reStarBold.ReplaceAllStringFunc(text, func(s string) string {
		// Preserve leading/trailing whitespace
		trimmed := strings.TrimSpace(s)
		if len(trimmed) >= 2 && trimmed[0] == '*' && trimmed[len(trimmed)-1] == '*' {
			inner := trimmed[1 : len(trimmed)-1]
			prefix := s[:strings.Index(s, "*")]
			suffix := s[strings.LastIndex(s, "*")+1:]
			return prefix + "\033[1m" + inner + "\033[22m" + suffix
		}
		return s
	})

	// h1. h2. etc headers → just bold text
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 3 && trimmed[0] == 'h' && trimmed[1] >= '1' && trimmed[1] <= '6' && trimmed[2] == '.' {
			lines[i] = "\033[1m" + strings.TrimSpace(trimmed[3:]) + "\033[22m"
		}
	}

	return strings.Join(lines, "\n")
}
