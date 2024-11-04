package parser

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/twolodzko/datalogo/datalog"
)

func TestReadToken(t *testing.T) {
	var testCases = []struct {
		input, expected string
	}{
		{"foo", "foo"},
		{"42", "42"},
		{"-356 ", "-356"},
		{"foo(a, X).", "foo"},
		{"Yeti, Zombie", "Yeti"},
		{"B,C,D", "B"},
		{",C,D", ","},
		{"  = X, foo(bar),", "="},
		{`"hello, world!"), bar(abc, `, `"hello, world!"`},
		{"_, X", "_"},
		{"foo_bar, ", "foo_bar"},
		{"), foo(a,", ")"},
		{":- foo(A),", ":-"},
		{"<= 84", "<="},
	}

	for _, tt := range testCases {
		parser := NewParser(strings.NewReader(tt.input))

		token, err := parser.readToken()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
			continue
		}
		if token != tt.expected {
			t.Errorf(
				"for input '%s' expected '%s', got '%s'",
				tt.input, tt.expected, token,
			)
		}
	}
}

func TestReadTerm(t *testing.T) {
	var testCases = []struct {
		input    string
		expected any
	}{
		{
			"foo",
			String("foo"),
		},
		{
			"X",
			Var{Name: "X"},
		},
		{
			"0",
			0,
		},
		{
			"42",
			42,
		},
		{
			"+8",
			8,
		},
		{
			"-615",
			-615,
		},
		{
			`"hello, world!"`,
			String("hello, world!"),
		},
		{
			`""`,
			String(""),
		},
		{
			"2=2",
			2,
		},
	}

	for _, tt := range testCases {
		parser := NewParser(strings.NewReader(tt.input))
		result, err := parser.readTerm()
		if err != nil {
			t.Errorf("parsing '%s' thrown an error: %s", tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("for '%s' expected '%v', got '%v'", tt.input, tt.expected, result)
		}
	}
}

func TestReadLiteral(t *testing.T) {
	var testCases = []struct {
		input    string
		expected Evaluable
	}{
		{
			"foo(a, b)",
			Atom{
				Name: "foo",
				Args: []any{
					String("a"),
					String("b"),
				},
			},
		},
		{
			"X = Y,",
			Constraint{
				Op:  "=",
				Lhs: Var{Name: "X"},
				Rhs: Var{Name: "Y"},
			},
		},
		{
			"X != Y,",
			Constraint{
				Op:  "!=",
				Lhs: Var{Name: "X"},
				Rhs: Var{Name: "Y"},
			},
		},
		{
			"1 != 1,",
			Constraint{
				Op:  "!=",
				Lhs: 1,
				Rhs: 1,
			},
		},
		{
			"51 >= 42,",
			Constraint{
				Op:  ">=",
				Lhs: 51,
				Rhs: 42,
			},
		},
		{
			`"aaaa" < "bbbb")`,
			Constraint{
				Op:  "<",
				Lhs: String("aaaa"),
				Rhs: String("bbbb"),
			},
		},
	}

	for _, tt := range testCases {
		parser := NewParser(strings.NewReader(tt.input))
		result, err := parser.readLiteral()
		if err != nil {
			t.Errorf("parsing '%s' thrown an error: %s", tt.input, err)
			continue
		}
		if !cmp.Equal(result, tt.expected) {
			t.Errorf("for '%s' expected '%v', got '%v'", tt.input, tt.expected, result)
		}
	}
}

func TestParser(t *testing.T) {
	var testCases = []struct {
		input    string
		expected any
	}{
		{
			"foo(a).",
			Assertion{
				Fact: Atom{
					Name: "foo",
					Args: []any{
						String("a"),
					},
				},
			},
		},
		{
			"bar(X, b).",
			Assertion{
				Fact: Atom{
					Name: "bar",
					Args: []any{
						Var{Name: "X"},
						String("b"),
					},
				},
			},
		},
		{
			"baz(A, B) :- foo(A), bar(A, B).",
			Assertion{
				Fact: Rule{
					Atom: Atom{
						Name: "baz",
						Args: []any{
							Var{Name: "A"},
							Var{Name: "B"},
						},
					},
					Body: []Evaluable{
						Atom{
							Name: "foo",
							Args: []any{
								Var{Name: "A"},
							},
						},
						Atom{
							Name: "bar",
							Args: []any{
								Var{Name: "A"},
								Var{Name: "B"},
							},
						},
					},
				},
			},
		},
		{
			"baz(X, Y)?",
			Query{
				Query: Atom{
					Name: "baz",
					Args: []any{
						Var{Name: "X"},
						Var{Name: "Y"},
					},
				},
			},
		},
		// parse a rule containing the constraints
		// this will re-order the terms to put the constraints on the back
		{
			"baz(A, B) :- A != B, foo(A).",
			Assertion{
				Fact: Rule{
					Atom: Atom{
						Name: "baz",
						Args: []any{
							Var{Name: "A"},
							Var{Name: "B"},
						},
					},
					Body: []Evaluable{
						Atom{
							Name: "foo",
							Args: []any{
								Var{Name: "A"},
							},
						},
						Constraint{
							Op:  "!=",
							Lhs: Var{Name: "A"},
							Rhs: Var{Name: "B"},
						},
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		parser := NewParser(strings.NewReader(tt.input))
		result, err := parser.Next()
		if err != nil {
			t.Errorf("parsing '%s' thrown an error: %s", tt.input, err)
			continue
		}
		if !cmp.Equal(result, tt.expected) {
			t.Errorf("for '%s' expected '%v', got '%v'", tt.input, tt.expected, result)
		}
	}
}
