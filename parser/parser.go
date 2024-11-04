package parser

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"

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
		return Assertion{Fact: atom}, nil
	case "?":
		return Query{Query: atom}, nil
	case "~":
		return Retraction{Fact: atom}, nil
	case ":-":
		body, err := p.readBody()
		if err != nil {
			return nil, err
		}
		rule := Rule{
			Atom: atom,
			Body: body,
		}
		return Assertion{Fact: rule}, nil
	default:
		return nil, UnexpectedToken{token}
	}
}

func (p *Parser) readLiteral() (Evaluable, error) {
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
		return Atom{
			Name: first,
			Args: args,
		}, err
	case isOperator(next):
		lhs, err := parseTerm(first)
		if err != nil {
			return nil, err
		}
		rhs, err := p.readTerm()
		return Constraint{
			Op:  next,
			Rhs: rhs,
			Lhs: lhs,
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

func (p *Parser) readBody() ([]Evaluable, error) {
	var body []Evaluable
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
			optimizeBody(body)
			return body, nil
		default:
			return nil, UnexpectedToken{token}
		}
	}
}

func optimizeBody(body []Evaluable) {
	// re-order the body to put the constraints at the back
	sort.Slice(body, func(i, j int) bool {
		if _, ok := body[i].(Atom); ok {
			_, ok := body[j].(Constraint)
			return ok
		}
		return false
	})
}

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
		return String(""), nil
	}
	switch {
	case isIdentifier(token):
		return String(token), nil
	case isVariable(token):
		return Var{Name: token}, nil
	case token == "_":
		return Wildcard{}, nil
	case isNumber(token):
		integer, err := strconv.Atoi(token)
		if err == nil {
			return integer, nil
		}
		return String(token), nil
	case token[0] == '"':
		end := len(token) - 1
		if end == 0 || token[end] != '"' {
			return nil, fmt.Errorf("invalid string: '%s'", token)
		}
		str := token[1:end]
		return String(str), nil
	default:
		return nil, UnexpectedToken{token}
	}
}

func isIdentifier(token string) bool {
	return 'a' <= token[0] && token[0] <= 'z'
}

func isVariable(token string) bool {
	return 'A' <= token[0] && token[0] <= 'Z'
}

func isNumber(token string) bool {
	return ('0' <= token[0] && token[0] <= '9') || token[0] == '-' || token[0] == '+'
}

func isOperator(token string) bool {
	switch token {
	case "=", "!=", "<", "<=", ">", ">=", "in":
		return true
	default:
		return false
	}
}

type UnexpectedToken struct {
	token string
}

func (err UnexpectedToken) Error() string {
	return fmt.Sprintf("unexpected token: '%s'", err.token)
}
