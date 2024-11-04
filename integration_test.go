package main_test

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"testing"

	. "github.com/twolodzko/datalogo/datalog"
	"github.com/twolodzko/datalogo/eval"
	"github.com/twolodzko/datalogo/parser"
)

func TestUnify(t *testing.T) {
	var testCases = []struct {
		lhs, rhs any
		vars     Vars
		expected bool
	}{
		// constants
		{String("wrong"), String("invalid"), Vars{}, false},
		{String("ok"), String("ok"), Vars{}, true},
		// wildcards
		{String("ok"), Wildcard{}, Vars{}, true},
		{Wildcard{}, String("ok"), Vars{}, true},
		{Var{Name: "X"}, Wildcard{}, Vars{}, true},
		{Wildcard{}, Var{Name: "Y"}, Vars{}, true},
		// variables
		{Var{Name: "X"}, Var{Name: "Y"}, Vars{}, true},
		{Var{Name: "X"}, Var{Name: "Y"}, newVars(
			Mapping{Var{Name: "X"}, String("ok")},
		), true},
		{Var{Name: "X"}, Var{Name: "Y"}, newVars(
			Mapping{Var{Name: "Y"}, String("ok")},
		), true},
		{Var{Name: "X"}, Var{Name: "Y"}, newVars(
			Mapping{Var{Name: "X"}, String("ok")},
			Mapping{Var{Name: "Y"}, String("ok")},
		), true},
		{Var{Name: "X"}, Var{Name: "Y"}, newVars(
			Mapping{Var{Name: "X"}, String("wrong")},
			Mapping{Var{Name: "Y"}, String("invalid")},
		), false},
	}

	for i, tt := range testCases {
		result := tt.vars.Unify(tt.lhs, tt.rhs)
		if result != tt.expected {
			t.Errorf(
				"test case %d: %v = %v was expected to be %v, got %v (vars: %v)",
				i+1, tt.lhs, tt.rhs, tt.expected, result, tt.vars,
			)
		}
	}
}

func TestUnifyComplex(t *testing.T) {
	vars := newVars(
		Mapping{Var{Name: "X"}, String("ok")},
		Mapping{Var{Name: "Y"}, Var{Name: "X"}},
		Mapping{Var{Name: "Z"}, Var{Name: "Y"}},
	)

	// unify with existing variables
	if !vars.Unify(Var{Name: "Z"}, String("ok")) {
		t.Errorf("failed to unify with a constant")
	}
	if !vars.Unify(Var{Name: "Z"}, Var{Name: "X"}) {
		t.Errorf("failed to unify variables")
	}

	// unify new variable
	if !vars.Unify(Var{Name: "N"}, String("new")) {
		t.Errorf("failed to unify new variable")
	}
	if val, _ := vars.Get(Var{Name: "N"}); val != String("new") {
		t.Errorf("new variable is not present")
	}

	// negative case
	if vars.Unify(Var{Name: "N"}, Var{Name: "Y"}) {
		t.Errorf("the new variable should not unify with the previous one")
	}
}

func TestUnifyPropagates(t *testing.T) {
	vars := newVars(
		Mapping{Var{Name: "B"}, Var{Name: "A"}},
		Mapping{Var{Name: "C"}, Var{Name: "B"}},
		Mapping{Var{Name: "D"}, Var{Name: "A"}},
	)

	if !vars.Unify(Var{Name: "B"}, String("ok")) {
		t.Errorf("initial unification failed")
	}
	// all of them should "have" the value
	for _, v := range "ABCD" {
		if !vars.Unify(Var{Name: string(v)}, String("ok")) {
			t.Errorf("unification against '%v' failed", v)
		}
	}
}

func TestQuery(t *testing.T) {
	db := newDatabaseFrom(
		// no match: wrong arity
		Atom{
			Name: "baz",
			Args: []any{String("a")},
		},
		Atom{
			Name: "foo",
			Args: []any{String("wrong"), String("wrong"), String("wrong"), String("wrong")},
		},
		Atom{
			Name: "foo",
			Args: []any{String("wrong"), String("invalid"), String("bad")},
		},
		// no match: wrong name
		Atom{
			Name: "bar",
			Args: []any{String("a"), String("wrong")},
		},
		// matches
		Atom{
			Name: "foo",
			Args: []any{Var{Name: "Y"}, Var{Name: "Y"}},
		},
		Atom{
			Name: "foo",
			Args: []any{String("a"), String("b")},
		},
		Atom{
			Name: "foo",
			Args: []any{String("a"), String("c")},
		},
		Atom{
			Name: "foo",
			Args: []any{Var{Name: "Y"}, String("d")},
		},
		Rule{
			Atom: Atom{
				Name: "foo",
				Args: []any{Var{Name: "Y"}, String("e")},
			},
			Body: []Evaluable{
				// this isn't really a condition, just for a trivial test case
				Atom{
					Name: "baz",
					Args: []any{String("a")},
				},
			},
		},
		// no match: wrong argument
		Atom{
			Name: "foo",
			Args: []any{String("wrong"), String("invalid")},
		},
	)

	// foo(a, X)?
	query := Query{
		Query: Atom{
			Name: "foo",
			Args: []any{String("a"), Var{Name: "X"}},
		},
	}

	results, err := evalAndCollect(query, db)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if len(results) != 5 {
		t.Errorf("wrong number of results in: %v", results)
		return
	}

	// make sure they are sorted so we can run the test below
	sort.Slice(results, func(i, j int) bool {
		lhs, _ := results[i].Args[1].(String)
		rhs, _ := results[j].Args[1].(String)
		return lhs < rhs
	})

	for i, id := range "abcde" {
		result := results[i].Args[1]
		if result != String(id) {
			t.Errorf("in env[%d] expected: %v, got %v", i, string(id), result)
		}
	}
}

func TestQueryRule(t *testing.T) {
	db := newDatabaseFrom(
		// foo(a).
		Atom{
			Name: "foo",
			Args: []any{String("a")},
		},
		// bar(X, b) :- foo(X).
		Rule{
			Atom: Atom{
				Name: "bar",
				Args: []any{Var{Name: "X"}, String("b")},
			},
			Body: []Evaluable{
				Atom{
					Name: "foo",
					Args: []any{Var{Name: "X"}},
				},
			},
		},
	)

	// bar(A, B)?
	query := Query{
		Query: Atom{
			Name: "bar",
			Args: []any{Var{Name: "A"}, Var{Name: "B"}},
		},
	}

	results, err := evalAndCollect(query, db)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if len(results) != 1 {
		t.Errorf("wrong number of results in: %v", results)
		return
	}

	expected := Atom{
		Name: "bar",
		Args: []any{
			String("a"), String("b"),
		},
	}
	if !results[0].Equal(expected) {
		t.Errorf("expected: %v, got: %v", expected, results[0])
	}
}

func TestQueryLongerRule(t *testing.T) {
	db := newDatabaseFrom(
		// foo(ok).
		Atom{
			Name: "foo",
			Args: []any{String("ok")},
		},
		// same(Z, Z).
		Atom{
			Name: "same",
			Args: []any{Var{Name: "Z"}, Var{Name: "Z"}},
		},
		// bar(X, Y) :- same(X, Y), foo(X).
		Rule{
			Atom: Atom{
				Name: "bar",
				Args: []any{Var{Name: "X"}, Var{Name: "Y"}},
			},
			Body: []Evaluable{
				Atom{
					Name: "same",
					Args: []any{Var{Name: "X"}, Var{Name: "Y"}},
				},
				Atom{
					Name: "foo",
					Args: []any{Var{Name: "X"}},
				},
			},
		},
	)

	// bar(A, B)?
	query := Query{
		Query: Atom{
			Name: "bar",
			Args: []any{Var{Name: "A"}, Var{Name: "B"}},
		},
	}

	results, err := evalAndCollect(query, db)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if len(results) != 1 {
		t.Errorf("wrong number of results in: %v", results)
		return
	}

	expected := Atom{
		Name: "bar",
		Args: []any{
			String("ok"), String("ok"),
		},
	}
	if !results[0].Equal(expected) {
		t.Errorf("expected: %v, got: %v", expected, results[0])
	}
}

func TestIntegration(t *testing.T) {
	var testCases = []struct {
		input    string
		expected []Atom
	}{
		{
			`
			human(socrates).
			mortal(X) :- human(X).
			mortal(socrates)?
			`,
			[]Atom{
				{
					Name: "mortal",
					Args: []any{
						String("socrates"),
					},
				},
			},
		},
		{
			`
			jump(X, 1) :- jump(X, 2).
			jump(X, 2) :- jump(X, 3).
			jump(X, 3) :- jump(X, 4).
			jump(X, 4) :- jump(X, 5).
			jump(done, 5).
			jump(X, 1)?
			`,
			[]Atom{
				{
					Name: "jump",
					Args: []any{
						String("done"),
						1,
					},
				},
			},
		},
		{
			`
			parent(xerces, brooke).
			parent(brooke, damocles).
			ancestor(X, Y) :- parent(X, Y).
			ancestor(X, Y) :- parent(X, Z), ancestor(Z, Y).
			ancestor(xerces, X)?
			`,
			[]Atom{
				{
					Name: "ancestor",
					Args: []any{
						String("xerces"),
						String("brooke"),
					},
				},
				{
					Name: "ancestor",
					Args: []any{
						String("xerces"),
						String("damocles"),
					},
				},
			},
		},
		{
			`
			edge(a, b).
			edge(b, c).
			edge(c, d).
		 	path(X, Y) :- edge(X, Y).
		 	path(X, Y) :- edge(X, Z), path(Z, Y).
		 	path(a, X)?
			`,
			[]Atom{
				{
					Name: "path",
					Args: []any{
						String("a"),
						String("b"),
					},
				},
				{
					Name: "path",
					Args: []any{
						String("a"),
						String("c"),
					},
				},
				{
					Name: "path",
					Args: []any{
						String("a"),
						String("d"),
					},
				},
			},
		},
		{
			`
			same(Z, Z).
			foo(A, B) :- same(A, B), bar(A, x).
			bar(a, _).
			foo(X, Y)?
			`,
			[]Atom{
				{
					Name: "foo",
					Args: []any{
						String("a"),
						String("a"),
					},
				},
			},
		},
		{
			// this stackoverflows if implemented incorrectly
			`
			foo(a,b).
			bar(X,Y) :- foo(X,Y).
			baz(X,Y) :- bar(Y,X).
			baz(B,A)?
			`,
			[]Atom{
				{
					Name: "baz",
					Args: []any{
						String("b"),
						String("a"),
					},
				},
			},
		},
		{
			`
			foo(ok).
			foo(wrong).
			foo(fine).
			foo(wrong)~
			foo(X)?
			`,
			[]Atom{
				{
					Name: "foo",
					Args: []any{
						String("fine"),
					},
				},
				{
					Name: "foo",
					Args: []any{
						String("ok"),
					},
				},
			},
		},
		// constraints
		{
			`
			less(A, B) :- A < B.
			less(1, 3)?
			`,
			[]Atom{
				{
					Name: "less", Args: []any{1, 3},
				},
			},
		},
		{
			`
			foo(a).
			foo(b).
			foo(c).
			bar(X) :- X != b, foo(X).
			bar(X)?
			`,
			[]Atom{
				{
					Name: "bar",
					Args: []any{
						String("a"),
					},
				},
				{
					Name: "bar",
					Args: []any{
						String("c"),
					},
				},
			},
		},
	}
	for _, tt := range testCases {
		db := make(Database)
		result, err := evalString(tt.input, db)

		sort.Slice(result, func(i, j int) bool {
			return result[i].String() < result[j].String()
		})

		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("for:\n%v\nexpected: %v, got: %v", tt.input, tt.expected, result)
		}
	}
}

func evalAndCollect(query any, db Database) ([]Atom, error) {
	ch := make(chan Atom)
	if err := eval.Eval(query, db, ch); err != nil {
		return nil, fmt.Errorf("unexpected error: %s", err)
	}
	var results []Atom
	for val := range ch {
		results = append(results, val)
	}
	return results, nil
}

func evalString(code string, db Database) ([]Atom, error) {
	var result []Atom
	parser := parser.NewParser(strings.NewReader(code))
	for {
		expr, err := parser.Next()
		if err != nil {
			if err == io.EOF {
				return result, nil
			}
			return nil, err
		}

		ch := make(chan Atom)
		if err = eval.Eval(expr, db, ch); err != nil {
			return nil, err
		}
		if _, ok := expr.(Query); ok {
			for val := range ch {
				result = append(result, val)
			}
		}
	}
}

func newDatabaseFrom(clauses ...HasKey) Database {
	db := make(Database)
	for _, clause := range clauses {
		db.Assert(clause)
	}
	return db
}

func newVars(mapping ...Mapping) Vars {
	return Vars{
		Mapping: mapping,
	}
}
