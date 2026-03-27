package datalog

import (
	"fmt"
)

// reduce constraints
// * any is replaced by whatever constraint
// * < is included in <=
// * < is included in <=
// * < X & = X --> <= X
// * = has lower priority than !=
// * CONST (< | <=) VARIABLE --> VARIABLE op CONST

// func (c Constraint) Eval(vars Vars, _ Database, ch chan<- Vars) {
// 	lhs := vars.expand(c.Lhs)
// 	rhs := vars.expand(c.Rhs)
// 	if c.evalWith(lhs, rhs) {
// 		ch <- vars
// 	}
// }

// func (c Constraint) simplify() any {
// 	if _, ok := c.Lhs.(Any); ok {
// 		return Any{}
// 	}
// 	if _, ok := c.Rhs.(Any); ok {
// 		return Any{}
// 	}
// 	if _, ok := c.Rhs.(Variable); ok {
// 		if _, ok := c.Lhs.(Variable); ok {
// 			return c
// 		}
// 		switch c.Op {
// 		case "<":
// 			c.Op = ">"
// 		case "<=":
// 			c.Op = ">="
// 		case ">":
// 			c.Op = "<"
// 		case ">=":
// 			c.Op = "<="
// 		}
// 		c.Lhs, c.Rhs = c.Rhs, c.Lhs
// 	}
// 	return c
// }

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
