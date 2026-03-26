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

type Variable string

type Atom struct {
	Name string
	Args []any
}

// TODO
type Match any

type Rule struct {
	Atom
	Body []Match
}

// Constraints are basic inequalities and equalities applied to primitive types.
// See: https://souffle-lang.github.io/constraints
// type Constraint struct {
// 	Op       string
// 	Lhs, Rhs any
// }

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

// func (c Constraint) String() string {
// 	return fmt.Sprintf("%v %s %v", c.Lhs, c.Op, c.Rhs)
// }

func (w Any) String() string {
	return "_"
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

func (a Atom) Eval(vars Vars, _ Database) {
	// TODO
}
