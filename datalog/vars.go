package datalog

// Variables substitutions mapping to be used during unification.
type Vars struct {
	Counter uint
	Mapping []Mapping
}

type Mapping struct {
	Key Var
	Val any
}

// Unify the two values and return the status.
// When unifying with variables, store the
// substitution.
func (v *Vars) Unify(lhs, rhs any) bool {
	if lhs == rhs {
		return true
	}
	if _, ok := lhs.(Wildcard); ok {
		return true
	}
	if _, ok := rhs.(Wildcard); ok {
		return true
	}
	if lhs, ok := lhs.(Var); ok {
		return v.unifyVar(lhs, rhs)
	}
	if rhs, ok := rhs.(Var); ok {
		return v.unifyVar(rhs, lhs)
	}
	return false
}

func (v *Vars) unifyVar(key Var, val any) bool {
	if key, ok := v.Get(key); ok {
		return v.Unify(key, val)
	}
	if newVal, ok := val.(Var); ok {
		if newVal, ok := v.Get(newVal); ok {
			val = newVal
		}
	}
	v.Mapping = append(v.Mapping, Mapping{
		Key: key,
		Val: val,
	})
	return true
}

func (vars Vars) unifyAll(lhs, rhs []any) (bool, Vars) {
	if len(lhs) != len(rhs) {
		return false, vars
	}
	for i := 0; i < len(lhs); i++ {
		if !vars.Unify(lhs[i], rhs[i]) {
			return false, vars
		}
	}
	return true, vars
}

// Get the most recent value associated with the key
// searching from the most recent substitution history
// backwards.
func (v Vars) Get(key Var) (any, bool) {
	for i := len(v.Mapping) - 1; i >= 0; i-- {
		if v.Mapping[i].Key == key {
			return v.Mapping[i].Val, true
		}
	}
	return nil, false
}

// Substitute all the variables with corresponding values.
func (v *Vars) substitute() {
	for last := len(v.Mapping) - 1; last > 0; last-- {
		new := v.Mapping[last]
		for i := last - 1; i >= 0; i-- {
			this := &v.Mapping[i]
			if this.Val == new.Key {
				(*this).Val = new.Val
			}
		}
	}
}

// "Rename" the variable to avoid name clashes
// in the substitutions table.
func (vars Vars) rename(val any) any {
	if val, ok := val.(Var); ok {
		val.Counter = vars.Counter
		return val
	}
	return val
}
