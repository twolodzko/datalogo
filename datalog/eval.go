package datalog

import "sync"

// Find all the facts in the database that unify with the query
// and send the matched variable substitutions to the out channel.
func (query Atom) Eval(vars Vars, db Database, out chan<- Vars) {
	var wg sync.WaitGroup
	ch := make(chan Evaluable)
	go db.find(query, ch)

	vars.Counter++
	for fact := range ch {
		wg.Add(1)
		go func() {
			defer wg.Done()
			query.unify(fact, vars, db, out)
		}()
	}
	wg.Wait()
}

// Unify the query with the fact. If the query is matched with
// Rules, evaluate them recursively. Send send the matched
// variable substitutions to the out channel.
func (query Atom) unify(fact any, vars Vars, db Database, out chan<- Vars) {
	switch fact := fact.(type) {
	case Atom:
		atom := fact.renameVars(vars)
		if ok, vars := vars.unifyAll(query.Args, atom.Args); ok {
			out <- vars
		}
	case Rule:
		rule := fact.renameVars(vars)
		if ok, vars := vars.unifyAll(query.Args, rule.Args); ok {
			evalBody(rule.Body, vars, db, out)
		}
	}
}

// Evaluate the body of the Rule, send the matched variable substitutions
// to the out channel.
func evalBody(body []Evaluable, vars Vars, db Database, out chan<- Vars) {
	if len(body) == 0 {
		// better than index error
		panic("rule's body cannot be empty")
	}

	ch := make(chan Vars)
	go func() {
		body[0].Eval(vars, db, ch)
		close(ch)
	}()

	for vars := range ch {
		if len(body) == 1 {
			out <- vars
		} else {
			evalBody(body[1:], vars, db, out)
		}
	}
}

func (a Atom) renameVars(vars Vars) Atom {
	var args []any
	for _, arg := range a.Args {
		arg = vars.rename(arg)
		args = append(args, arg)
	}
	return Atom{
		Name: a.Name,
		Args: args,
	}
}

func (r Rule) renameVars(vars Vars) Rule {
	atom := r.Atom.renameVars(vars)
	var body []Evaluable
	for _, lit := range r.Body {
		switch this := lit.(type) {
		case Atom:
			lit = this.renameVars(vars)
		case Rule:
			lit = this.renameVars(vars)
		case Constraint:
			lit = this.renameVars(vars)
		}
		body = append(body, lit)
	}
	return Rule{
		Atom: atom,
		Body: body,
	}
}

func (c Constraint) renameVars(vars Vars) Constraint {
	return Constraint{
		Op:  c.Op,
		Lhs: vars.rename(c.Lhs),
		Rhs: vars.rename(c.Rhs),
	}
}
