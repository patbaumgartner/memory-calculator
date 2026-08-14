// Package parser provides utilities for parsing flags and options.
package parser

import (
	"fmt"
	"strings"
	"unicode"
)

type quoteState uint8

const (
	unquoted quoteState = iota
	singleQuoted
	doubleQuoted
)

// ParseFlags splits a JVM options string into individual arguments using POSIX-like shell word
// rules: unquoted whitespace separates arguments, single quotes are literal, double quotes and
// unquoted text honor backslash escapes, and a quoted section joins the word around it rather
// than becoming a word of its own.
//
// Unterminated quotes and dangling escapes are rejected, because silently mis-splitting a JVM
// option string would produce memory flags the caller never asked for.
func ParseFlags(input string) ([]string, error) {
	var (
		result  []string
		current strings.Builder
		state   quoteState
		escaped bool
		started bool
	)

	for _, r := range input {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false

		case r == '\\' && state != singleQuoted:
			escaped = true
			started = true

		case isQuoteTransition(r, state):
			state = nextQuoteState(r, state)
			started = true

		case state == unquoted && unicode.IsSpace(r):
			if started {
				result = append(result, current.String())
				current.Reset()
				started = false
			}

		default:
			current.WriteRune(r)
			started = true
		}
	}

	if escaped {
		return nil, fmt.Errorf("unterminated escape sequence in %q", input)
	}

	if state != unquoted {
		return nil, fmt.Errorf("unterminated quote in %q", input)
	}

	if started {
		result = append(result, current.String())
	}

	return result, nil
}

// isQuoteTransition reports whether the rune opens or closes the current quoting context. A quote
// character inside the opposite quoting style is literal text, not a delimiter.
func isQuoteTransition(r rune, state quoteState) bool {
	switch state {
	case unquoted:
		return r == '\'' || r == '"'
	case singleQuoted:
		return r == '\''
	case doubleQuoted:
		return r == '"'
	}
	return false
}

func nextQuoteState(r rune, state quoteState) quoteState {
	if state != unquoted {
		return unquoted
	}
	if r == '\'' {
		return singleQuoted
	}
	return doubleQuoted
}
