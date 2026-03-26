package parser

import (
	"bufio"
	"fmt"
	"io"

	//lint:ignore ST1001 this is an internal dependency
	. "github.com/twolodzko/datalogo/datalog"
)

type Parser struct {
	*bufio.Reader
}

func NewParser(in io.Reader) *Parser {
	return &Parser{bufio.NewReader(in)}
}

func (p *Parser) Next() (any, error) {
	head, err := p.readToken()
	if err != nil {
		return nil, err
	}

	var atom Atom
	switch {
	case isIdentifier(head):
		if err = p.expect("("); err != nil {
			return nil, err
		}
		args, err := p.readArgs()
		if err != nil {
			return nil, err
		}
		atom = Atom{
			Name: head,
			Args: args,
		}
	case head == "#input":
		return p.readInput()
	default:
		return nil, UnexpectedToken{head}
	}

	token, err := p.readToken()
	if err != nil {
		return nil, err
	}

	switch token {
	case ".":
		return Assertion{atom}, nil
	case "?":
		return Question{atom}, nil
	case "~":
		return Retraction{atom}, nil
	case ":-":
		body, err := p.readBody()
		if err != nil {
			return nil, err
		}
		rule := Rule{
			Atom: atom,
			Body: body,
		}
		return Assertion{rule}, nil
	default:
		return nil, UnexpectedToken{token}
	}
}

func (p *Parser) readLiteral() (Match, error) {
	first, err := p.readToken()
	if err != nil {
		return nil, err
	}
	next, err := p.readToken()
	if err != nil {
		return nil, err
	}
	switch {
	case next == "(":
		args, err := p.readArgs()
		if err != nil {
			return nil, err
		}
		if isOperator(first) && len(args) != 2 {
			return nil, fmt.Errorf("%s has wrong number of arguments: %d != 2", first, len(args))
		}
		return Atom{
			Name: first,
			Args: args,
		}, nil
	case isOperator(next):
		lhs, err := parseTerm(first)
		if err != nil {
			return nil, err
		}
		rhs, err := p.readTerm()
		return Atom{
			Name: next,
			Args: []any{lhs, rhs},
		}, err
	default:
		return nil, UnexpectedToken{next}
	}
}

func (p *Parser) expect(expected string) error {
	token, err := p.readToken()
	if err != nil {
		return err
	}
	if token != expected {
		return UnexpectedToken{token}
	}
	return nil
}

func (p *Parser) readBody() ([]Match, error) {
	var body []Match
	for {
		atom, err := p.readLiteral()
		if err != nil {
			return nil, err
		}
		body = append(body, atom)

		token, err := p.readToken()
		if err != nil {
			return nil, err
		}
		switch token {
		case ",", "&":
			// expected
		case ".":
			// optimizeBody(body)
			return body, nil
		default:
			return nil, UnexpectedToken{token}
		}
	}
}

// func optimizeBody(body []Match) {
// 	// re-order the body to put the constraints at the back
// 	sort.Slice(body, func(i, j int) bool {
// 		if _, ok := body[i].(Atom); ok {
// 			_, ok := body[j].(Constraint)
// 			return ok
// 		}
// 		return false
// 	})
// }

func (p *Parser) readArgs() ([]any, error) {
	var args []any
	for {
		term, err := p.readTerm()
		if err != nil {
			return nil, err
		}
		args = append(args, term)

		token, err := p.readToken()
		if err != nil {
			return nil, err
		}
		switch token {
		case ",":
			// expected
		case ")":
			return args, nil
		default:
			return nil, UnexpectedToken{token}
		}
	}
}

func (p *Parser) readTerm() (any, error) {
	token, err := p.readToken()
	if err != nil {
		return nil, err
	}
	return parseTerm(token)
}

func parseTerm(token string) (any, error) {
	if len(token) == 0 {
		return "", nil
	}
	switch {
	case isVariable(token):
		return Variable(token), nil
	case token == "_":
		return Any{}, nil
	case token[0] == '"':
		end := len(token) - 1
		if end == 0 || token[end] != '"' {
			return nil, fmt.Errorf("invalid string: '%s'", token)
		}
		str := token[1:end]
		return str, nil
	default:
		return token, nil
	}
}

func isIdentifier(token string) bool {
	return 'a' <= token[0] && token[0] <= 'z'
}

func isVariable(token string) bool {
	return 'A' <= token[0] && token[0] <= 'Z'
}

func isOperator(token string) bool {
	switch token {
	case "=", "!=", "<", "<=", ">", ">=":
		return true
	default:
		return false
	}
}

type UnexpectedToken struct {
	token string
}

func (err UnexpectedToken) Error() string {
	return fmt.Sprintf("unexpected: '%s'", err.token)
}
