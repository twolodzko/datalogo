package datalog

import (
	"fmt"
	"strconv"
	"strings"
)

func (c Constraint) Eval(vars Vars, _ Database, ch chan<- Vars) {
	lhs := vars.expand(c.Lhs)
	rhs := vars.expand(c.Rhs)
	if c.evalWith(lhs, rhs) {
		ch <- vars
	}
}

// Check if the constraint holds for the arguments.
func (c Constraint) evalWith(lhs, rhs any) bool {
	if c.Op == "in" {
		if lhs, ok := toString(lhs); ok {
			if rhs, ok := toString(rhs); ok {
				return strings.Contains(rhs, lhs)
			}
		}
		return false
	}
	if lhs, rhs, ok := asType[String](lhs, rhs); ok {
		if compare(c.Op, lhs, rhs) {
			return true
		}
	}
	if lhs, rhs, ok := asType[int](lhs, rhs); ok {
		if compare(c.Op, lhs, rhs) {
			return true
		}
	}
	return false
}

func compare[T Const](op string, lhs, rhs T) bool {
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

func asType[T any](lhs, rhs any) (T, T, bool) {
	switch lhs := lhs.(type) {
	case T:
		rhs, ok := rhs.(T)
		return lhs, rhs, ok
	default:
		// see: https://stackoverflow.com/a/70589302
		return *new(T), *new(T), false
	}
}

// If key is a variable and has a value, return the value, otherwise return it.
func (vars Vars) expand(key any) any {
	for {
		switch k := key.(type) {
		case Var:
			if val, ok := vars.Get(k); ok {
				key = val
			} else {
				return key
			}
		default:
			return key
		}
	}
}

func toString(val any) (string, bool) {
	switch val := val.(type) {
	case String:
		return string(val), true
	case int:
		return strconv.Itoa(val), true
	default:
		return "", false
	}
}
