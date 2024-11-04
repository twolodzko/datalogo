package datalog

import (
	"reflect"
	"slices"
	"sync"
)

// Node holds data in a tree structure, where the path
// to the final node are the arguments of the Atom
// or the Rule, and the final node as a Value contains
// the stored object itself.
//
//	 foo(a, b, c).
//	 foo(a, X).
//	 foo(a, b, d).
//
//		         a
//		        / \
//		       b   ?
//		      / \   \
//		     c   d  foo(a,X)
//		    /     \
//	 foo(a,b,c)  foo(a,b,d)
type Node struct {
	Value any
	Next  []*Node
}

func nodeFrom(args []any, val any) *Node {
	if len(args) == 0 {
		return &Node{Value: val}
	}
	return &Node{
		Value: args[0],
		Next: []*Node{
			nodeFrom(args[1:], val),
		},
	}
}

// Recursively traverse the tree (or create new nodes)
// to add the value to it. Replace the Var arguments with Wildcards.
func (n *Node) add(args []any, val any) bool {
	var arg any
	switch args[0].(type) {
	case Wildcard, Var:
		// storage optimization: it doesn't matter what variable it is,
		// we need to unify it later
		arg = Wildcard{}
	default:
		arg = args[0]
	}

	if n.Value != arg {
		return false
	}

	if len(args) == 1 {
		if !slices.ContainsFunc(n.Next, func(elem *Node) bool {
			return reflect.DeepEqual(elem.Value, val)
		}) {
			n.Next = append(n.Next, &Node{Value: val})
		}
	} else {
		for i := 0; i < len(n.Next); i++ {
			if n.Next[i].add(args[1:], val) {
				return true
			}
		}
		n.Next = append(n.Next, nodeFrom(args[1:], val))
	}
	return true
}

// Traverse the tree and remove the value if it exists.
func (n *Node) remove(args []any, val Atom) {
	if !maybeUnifies(n.Value, args[0]) {
		return
	}
	if len(args) == 1 {
		for i, node := range n.Next {
			if this, ok := node.Value.(Atom); ok {
				if this.Equal(val) {
					n.Next = delete(n.Next, i)
					return
				}
			}
		}
	} else {
		for _, node := range n.Next {
			node.remove(args[1:], val)
		}
	}
}

// Find all the values that match the arguments path
// and send them to the out channel.
func (n Node) find(args []any, out chan<- Evaluable) {
	if len(args) == 0 {
		// final node
		switch val := n.Value.(type) {
		case Atom:
			out <- val
		case Rule:
			out <- val
		}
	} else {
		if maybeUnifies(args[0], n.Value) {
			var wg sync.WaitGroup
			for _, next := range n.Next {
				wg.Add(1)
				go func() {
					defer wg.Done()
					next.find(args[1:], out)
				}()
			}
			wg.Wait()
		}
	}
}

// The values are equal constants, or one of them is variable.
func maybeUnifies(lhs, rhs any) bool {
	switch lhs.(type) {
	case Var, Wildcard:
		return true
	}
	switch rhs.(type) {
	case Var, Wildcard:
		return true
	}
	return rhs == lhs
}

// Delete i-th element from the slice s. It does not preserve
// the slice order. See: https://stackoverflow.com/a/37335777
func delete[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	s = s[:len(s)-1]
	return s
}
