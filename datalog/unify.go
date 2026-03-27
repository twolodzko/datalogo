package datalog

// Variables substitutions mapping to be used during unification.
type Vars struct{ Mapping []Mapping }

type Mapping struct {
	Key Variable
	Val any
}

// Unify the two values and return the status.
// When unifying with variables, store the
// substitution.
func (v *Vars) Unify(lhs, rhs any) bool {
	if key, ok := lhs.(Variable); ok {
		// TODO: use birth records to optimize it
		lhs, _ = v.Get(key)
	}
	if key, ok := rhs.(Variable); ok {
		rhs, _ = v.Get(key)
	}
	if lhs == rhs {
		return true
	}
	if _, ok := lhs.(Any); ok {
		return true
	}
	if _, ok := rhs.(Any); ok {
		return true
	}
	if lhs, ok := lhs.(Variable); ok {
		v.Mapping = append(v.Mapping, Mapping{lhs, rhs})
		return true
	}
	if rhs, ok := rhs.(Variable); ok {
		v.Mapping = append(v.Mapping, Mapping{rhs, lhs})
		return true
	}
	return false
}

// Materialize the unified variable
func (v *Vars) Reify(key any) (string, bool) {
	if key, ok := key.(Variable); ok {
		return v.reify(key, 0)
	}
	val, ok := key.(string)
	return val, ok
}

func (v *Vars) reify(key Variable, start int) (string, bool) {
	for i := start; i < len(v.Mapping); i++ {
		if v.Mapping[i].Key == key {
			val := v.Mapping[i].Val
			if val, ok := val.(Variable); ok {
				val, ok := v.reify(val, i+1)
				if ok {
					// so we don't repeat the search next time
					v.Mapping[i].Val = val
				}
				return val, ok
			} else {
				return "", false
			}
		}
	}
	return "", false
}

func (vars Vars) UnifyAll(lhs, rhs []any) (bool, Vars) {
	if len(lhs) != len(rhs) {
		return false, vars
	}
	for i := range len(lhs) {
		if !vars.Unify(lhs[i], rhs[i]) {
			return false, vars
		}
	}
	return true, vars
}

// Get the most recent value associated with the key
// searching from the most recent substitution history
// backwards.
func (v Vars) Get(key Variable) (any, bool) {
	for i := len(v.Mapping) - 1; i >= 0; i-- {
		if v.Mapping[i].Key == key {
			return v.Mapping[i].Val, true
		}
	}
	return key, false
}
