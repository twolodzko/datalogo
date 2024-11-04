package datalog

import (
	"fmt"
	"sync"
)

type (
	Key      string
	Database map[Key][]*Node
)

// Assert (save) the value to the database.
func (db Database) Assert(val HasKey) {
	var args []any
	switch val := val.(type) {
	case Atom:
		args = val.Args
	case Rule:
		args = val.Args
	default:
		panic(fmt.Sprintf("%v has invalid type", val))
	}

	key := val.Key()
	nodes := db[key]
	for _, node := range nodes {
		if node.add(args, val) {
			return
		}
	}
	db[key] = append(nodes, nodeFrom(args, val))
}

// Query the database to find all the matches for the query.
// Return all the matches by sending them to the out channel.
func (db Database) Query(query Atom, out chan<- Atom) {
	ch := make(chan Vars)
	go func() {
		defer close(ch)
		query.Eval(Vars{}, db, ch)
	}()

	// post-process
	go func() {
		defer close(out)
		for vars := range ch {
			vars.substitute()
			atom := query.Materialize(vars)
			out <- atom
		}
	}()
}

// Find the potential (un-unified) matches to the query,
// send them to the out channel.
func (db Database) find(query Atom, out chan<- Evaluable) {
	defer close(out)
	key := query.Key()
	if nodes, ok := db[key]; ok {
		var wg sync.WaitGroup
		for _, node := range nodes {
			wg.Add(1)
			go func() {
				defer wg.Done()
				node.find(query.Args, out)
			}()
		}
		wg.Wait()
	}
}

// Remove the value from the database if it exists.
func (db Database) Remove(val Atom) {
	key := val.Key()
	nodes := db[key]
	for _, node := range nodes {
		node.remove(val.Args, val)
	}
	db[key] = nodes
}

// The value can be stored in a Database.
type HasKey interface {
	Key() Key
}

func (a Atom) Key() Key {
	return Key(a.Name)
}

func (r Rule) Key() Key {
	return r.Atom.Key()
}

// Fill-in the arguments which are variables with their
// realizations to create new Atom.
func (a Atom) Materialize(vars Vars) Atom {
	var args []any
	for _, arg := range a.Args {
		if v, ok := arg.(Var); ok {
			if val, ok := vars.Get(v); ok {
				arg = val
			}
		}
		args = append(args, arg)
	}
	return Atom{
		Name: a.Name,
		Args: args,
	}
}
