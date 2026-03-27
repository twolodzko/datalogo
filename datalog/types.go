package datalog

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// TODO
type Database struct {
	Facts map[string][][]string
	Rules map[string][]Rule
}

type Any struct{}

type Variable struct {
	Name string
	//
	Scope int
}

type Atom struct {
	Name string
	Args []any
}

type Rule struct {
	Atom
	Body []Literal
}

// Constraints are basic inequalities and equalities applied to primitive types.
// See: https://souffle-lang.github.io/constraints
type Constraint struct {
	Op       string
	Lhs, Rhs any
}

type Assertion struct {
	Fact any
}

type Retraction struct {
	Atom
}

type Question struct {
	Atom
}

func (lhs Atom) Equal(rhs Atom) bool {
	return lhs.Name == rhs.Name && slices.Equal(lhs.Args, rhs.Args)
}

func (a Atom) String() string {
	return fmt.Sprintf("%s(%v)", a.Name, stringify(a.Args))
}

func (r Rule) String() string {
	return fmt.Sprintf("%s(%v) :- %v", r.Name, stringify(r.Args), stringify(r.Body))
}

func (c Constraint) String() string {
	return fmt.Sprintf("%v %s %v", c.Lhs, c.Op, c.Rhs)
}

func (w Any) String() string {
	return "_"
}

func (v Variable) String() string {
	return v.Name
}

func stringify[T any](vals []T) string {
	var elems []string
	for _, val := range vals {
		elems = append(elems, fmt.Sprintf("%v", val))
	}
	return strings.Join(elems, ", ")
}

func isAlphanum(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !(unicode.IsLetter(r) || unicode.IsNumber(r)) {
			return false
		}
	}
	return true
}

type Literal interface {
	Query(Vars, Database) ([]Vars, bool)
}

func (a Atom) Query(vars Vars, db Database) ([]Vars, bool) {
	// TODO
	// for val in db.Find(a) { if vars.Unify(a, val) { out = append(out, ...)  }  }
	return nil, false
}

func (a Constraint) Query(vars Vars, _ Database) ([]Vars, bool) {
	lhs, ok := vars.Reify(a.Lhs)
	if !ok {
		return nil, false
	}
	rhs, ok := vars.Reify(a.Rhs)
	if !ok {
		return nil, false
	}
	return []Vars{vars}, compare(a.Op, lhs, rhs)
}

// Check if the constraint holds for the arguments.
func compare(op, lhs, rhs string) bool {
	switch op {
	case "=":
		return lhs == rhs
	case "!=":
		return lhs != rhs
	case "<":
		return lhs < rhs
	case "<=":
		return lhs <= rhs
	case ">":
		return lhs > rhs
	case ">=":
		return lhs >= rhs
	default:
		panic(fmt.Sprintf("invalid operator: %s", op))
	}
}
