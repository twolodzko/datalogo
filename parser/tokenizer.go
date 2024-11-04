package parser

import (
	"io"
	"strings"
	"unicode"
)

func (parser *Parser) readToken() (string, error) {
	var str strings.Builder
LOOP:
	for {
		r, _, err := parser.ReadRune()
		if err != nil {
			if err == io.EOF {
				if str.Len() == 0 {
					return "", err
				}
				break
			}
			return "", err
		}

		if unicode.IsSpace(r) {
			if str.Len() > 0 {
				break
			} else {
				continue
			}
		}

		switch r {
		case '.', '?', '~', '(', ')', '=', ',', '&':
			if str.Len() == 0 {
				str.WriteRune(r)
			} else {
				// unread, this is a next token
				if err = parser.UnreadRune(); err != nil {
					return "", err
				}
			}
			break LOOP
		case ':':
			if str.Len() == 0 {
				str.WriteRune(r)
				if err := parser.maybeRead('-', &str); err != nil {
					return "", err
				}
			} else {
				// unread, this is a next token
				if err = parser.UnreadRune(); err != nil {
					return "", err
				}
			}
			break LOOP
		case '<', '>', '!':
			if str.Len() == 0 {
				str.WriteRune(r)
				if err := parser.maybeRead('=', &str); err != nil {
					return "", err
				}
			} else {
				// unread, this is a next token
				if err = parser.UnreadRune(); err != nil {
					return "", err
				}
			}
			break LOOP
		case '%':
			if err := parser.skipLine(); err != nil {
				return "", err
			}
			continue
		case '"':
			if str.Len() == 0 {
				return parser.readString()
			} else {
				// unread, this is a next token
				if err = parser.UnreadRune(); err != nil {
					return "", err
				}
			}
			break LOOP
		default:
			// TODO: should I check if it is alphanumerical?
			str.WriteRune(r)
		}
	}
	return str.String(), nil
}

func (parser *Parser) maybeRead(expected rune, str *strings.Builder) error {
	r, _, err := parser.ReadRune()
	if err != nil && err != io.EOF {
		return err
	}
	if r == expected {
		str.WriteRune(r)
	} else {
		// unread, this is a next token
		if err := parser.UnreadRune(); err != nil {
			return err
		}
	}
	return nil
}

func (parser *Parser) readString() (string, error) {
	var str strings.Builder
	str.WriteRune('"')

	for {
		r, _, err := parser.ReadRune()
		if err != nil {
			if err == io.EOF {
				if str.Len() == 1 {
					return "", err
				}
				return str.String(), nil
			}
			return "", err
		}
		str.WriteRune(r)

		switch r {
		case '"':
			return str.String(), nil
		case '\\':
			r, _, err = parser.ReadRune()
			if err != nil {
				return "", err
			}
			str.WriteRune(r)
		}
	}
}

func (parser *Parser) skipLine() error {
	for {
		r, _, err := parser.ReadRune()
		if r == '\n' || err != nil {
			return err
		}
	}
}
