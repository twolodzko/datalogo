package datalog

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDbAssert(t *testing.T) {
	db := make(Database)

	// add a value
	first := Atom{
		Name: "foo",
		Args: []any{1, 2},
	}
	db.Assert(first)

	expected1 := Database{
		"foo": []*Node{
			{
				Value: 1,
				Next: []*Node{
					{
						Value: 2,
						Next: []*Node{
							{
								Value: first,
							},
						},
					},
				},
			},
		},
	}

	if !cmp.Equal(db, expected1) {
		t.Errorf("expected: %v, got: %v", expected1, db)
	}

	// no-op, this value already exists
	second := Atom{
		Name: "foo",
		Args: []any{1, 2},
	}
	db.Assert(second)

	if !cmp.Equal(db, expected1) {
		t.Errorf("expected: %v, got: %v", expected1, db)
	}

	// add a new one
	third := Atom{
		Name: "foo",
		Args: []any{1, 2, 3},
	}
	db.Assert(third)

	expected3 := Database{
		"foo": []*Node{
			{
				Value: 1,
				Next: []*Node{
					{
						Value: 2,
						Next: []*Node{
							{
								Value: first,
							},
							{
								Value: 3,
								Next: []*Node{
									{
										Value: third,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if !cmp.Equal(db, expected3) {
		t.Errorf("expected: %v, got: %v", expected3, db)
	}

	// add one on a new branch
	fourth := Atom{
		Name: "foo",
		Args: []any{4, 5},
	}
	db.Assert(fourth)

	expected4 := Database{
		"foo": []*Node{
			{
				Value: 1,
				Next: []*Node{
					{
						Value: 2,
						Next: []*Node{
							{
								Value: first,
							},
							{
								Value: 3,
								Next: []*Node{
									{
										Value: third,
									},
								},
							},
						},
					},
				},
			},
			{
				Value: 4,
				Next: []*Node{
					{
						Value: 5,
						Next: []*Node{
							{
								Value: fourth,
							},
						},
					},
				},
			},
		},
	}

	if !cmp.Equal(db, expected4) {
		t.Errorf("expected: %v, got: %v", expected4, db)
	}
}
