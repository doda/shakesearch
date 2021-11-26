package main

import (
	"fmt"
	"regexp"
	"strings"

	snowballeng "github.com/kljensen/snowball/english"
)

func extractTokens(text string) []string {
	// Process text via tokenizing, lowercasing, filtering
	// stop words and stemming
	tokens := regexp.MustCompile(`\w+`).FindAllString(text, -1)

	result := []string{}
	for _, token := range tokens {
		// lower it
		token = strings.ToLower(token)
		// filter out stop words
		if !stopWords[token] {
			// stem it
			token = snowballeng.Stem(token, false)
			result = append(result, token)
		}
	}
	return result
}

func formatLine(line, token string, lineno int) string {
	// Try and find the token in the line, bold that word and prepend the line number
	// Warning: Hack, this is not reliable or elegant by any means
	sampleRegexp := regexp.MustCompile(`(?i)\b(` + token + `.*?)\b`)

	line = sampleRegexp.ReplaceAllString(line, "<strong>$1</strong>")
	return fmt.Sprintf("#%04d: %s", lineno, line)
}
