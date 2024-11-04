package eval

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	//lint:ignore ST1001 this is an internal dependency
	. "github.com/twolodzko/datalogo/datalog"
	"github.com/twolodzko/datalogo/parser"
)

// Evaluate the expression, if it is a Query type, send the
// results to the out channel, and close the channel afterwards.
func Eval(expr any, db Database, out chan Atom) error {
	if _, ok := expr.(Query); !ok {
		close(out)
	}
	switch expr := expr.(type) {
	case Assertion:
		if val, ok := expr.Fact.(HasKey); ok {
			db.Assert(val)
		} else {
			return fmt.Errorf("%v cannot be stored in database", expr.Fact)
		}
	case Retraction:
		db.Remove(expr.Fact)
	case Query:
		db.Query(expr.Query, out)
	case parser.Input:
		reader, err := NewInputReader(expr)
		if err != nil {
			return err
		}
		for {
			atom, err := reader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			db.Assert(atom)
		}
	default:
		return fmt.Errorf("invalid expression type: %t", expr)
	}
	return nil
}

type InputReader struct {
	parser.Input
	row int
	parser.Parser
}

func NewInputReader(inp parser.Input) (InputReader, error) {
	var reader *bufio.Reader
	switch inp.Source {
	case "stdin":
		reader = bufio.NewReader(os.Stdin)
	default:
		file, err := os.Open(inp.Source)
		if err != nil {
			return InputReader{}, err
		}
		reader = bufio.NewReader(file)
	}
	return InputReader{
		Input: inp,
		Parser: parser.Parser{
			Reader: reader,
		},
	}, nil
}

func (r InputReader) Next() (Atom, error) {
	for r.row < r.Skip {
		if _, _, err := r.ReadLine(); err != nil {
			return Atom{}, err
		}
		r.row++
	}
	line, err := r.ReadString('\n')
	if err != nil {
		return Atom{}, err
	}
	if strings.TrimSpace(line) == "" {
		return Atom{}, io.EOF
	}
	return r.ParseLine(line)
}
