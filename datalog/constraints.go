package datalog

import (
	"fmt"
)

// func (c Constraint) Eval(vars Vars, _ Database, ch chan<- Vars) {
// 	lhs := vars.expand(c.Lhs)
// 	rhs := vars.expand(c.Rhs)
// 	if c.evalWith(lhs, rhs) {
// 		ch <- vars
// 	}
// }

// Check if the constraint holds for the arguments.
func (c Constraint) compare(lhs, rhs string) bool {
	switch c.Op {
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
		panic(fmt.Sprintf("invalid operator: %s", c.Op))
	}
}

// // If key is a variable and has a value, return the value, otherwise return it.
// func (vars Vars) expand(key any) any {
// 	for {
// 		switch k := key.(type) {
// 		case Var:
// 			if val, ok := vars.Get(k); ok {
// 				key = val
// 			} else {
// 				return key
// 			}
// 		default:
// 			return key
// 		}
// 	}
// }
