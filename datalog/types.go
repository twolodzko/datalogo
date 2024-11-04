package datalog

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

type Const interface {
	~string | ~int
}

type (
	String   string
	Wildcard struct{}
)

type Var struct {
	Name    string
	Counter uint
}

type Atom struct {
	Name string
	Args []any
}

type Rule struct {
	Atom
	Body []Evaluable
}

// Constraints are basic inequalities and equalities applied to primitive types.
// See: https://souffle-lang.github.io/constraints
type Constraint struct {
	Op       string
	Lhs, Rhs any
}

type Evaluable interface {
	Eval(Vars, Database, chan<- Vars)
}

type Assertion struct {
	Fact any
}

type Retraction struct {
	Fact Atom
}

type Query struct {
	Query Atom
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

func (w Wildcard) String() string {
	return "_"
}

func (v Var) String() string {
	if v.Counter == 0 {
		return v.Name
	}
	return fmt.Sprintf("%s.%d", v.Name, v.Counter)
}

func (s String) String() string {
	str := string(s)
	if isAlphanum(str) {
		return str
	}
	return fmt.Sprintf("\"%s\"", str)
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
