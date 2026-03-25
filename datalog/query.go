package datalog

import "slices"

type Query struct {
	vars []Mapping
	body [][]QueryElement
}

type QueryElement struct {
	Atom
	constrains []Constraint
	negated    bool
}

// type QueryElement interface {
// 	Filter(Vars, Database) [][]string
// 	Cost() int
// }

func (a Atom) Matches(vars Vars, db Database) [][]string {
	return nil
}

func (a Atom) Cost() int {
	return 0
}

func (c Constraint) Matches(vars Vars, db Database) [][]string {
	return nil
}

func (c Constraint) Cost() int {
	return 0
}

func sortBody(elems []QueryElement) []QueryElement {
	slices.SortFunc(elems, func(a, b QueryElement) int {
		return b.Cost() - a.Cost()
	})
	return elems
}
