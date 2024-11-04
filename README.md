# Datalo.go

[Datalog] is a declarative logic programming language which syntactically is a subset of Prolog.

## Language tour

As explained in the Wikipedia's article on [Datalog]:

> A Datalog program consists of *facts*, which are statements that are held to be true, and *rules*,
> which say how to deduce new facts from known facts. For example, here are two facts that mean
> *xerces is a parent of brooke* and *brooke is a parent of damocles:*
>
> ```prolog
> parent(xerces, brooke).
> parent(brooke, damocles).
> ```
>
> The names are written in lowercase because strings beginning with an uppercase letter stand
> for variables. Here are two rules:
>
> ```prolog
> ancestor(X, Y) :- parent(X, Y).
> ancestor(X, Y) :- parent(X, Z), ancestor(Z, Y).
> ```

Rules consist of *heads* and *bodies* split by `:-`.

The facts and rules as arguments may take variables (their names start with uppercase letters),
wildcards `_` wchich are similar to variables but do not bind the values, and constants:
integers and strings. A string can either be an alphanumeric word starting with a lowercase letter,
like `xerces`, or a quoted string which can contain arbitrary characters `"Hello, world!"`.

The facts can be *asserted* (saved) to *database*:

```prolog
human(socrates).        % socrates is a human
mortal(X) :- human(X).  % X is mortal if X is a human
```

It is also possible to *remove* the fact from database:

```prolog
human(zeus)~
```

We can also *query* the database to answer our questions:

```prolog
mortal(socrates)?  % is socrates mortal?
mortal(Y)?         % list everyone (Y) who is mortal
```

The query would return a set of answers matching it.

## Database

The facts and rules are stored in a *database*, which is implemented
as a tree structure for an efficient retrieval.
For example, the following facts

```prolog
foo(a, b, c).
foo(a, X).
foo(a, b, d).
```

would be stored as

```text
          foo
            \
             a
            / \
           b   ?
          / \   \
         c   d  foo(a,X)
        /     \
foo(a,b,c)   foo(a,b,d)
```

In the diagram above the variable `X` is shown as `?` in the node's value
because both named variables and wildcards are represented the
same on the branches.

When the database is queried, the branches of the tree are traversed as
long as they are matching the arguments of the query. When the final node
is reached, the value stored in the node is *[unified]* with the query
and the result is returned only if it matches.

The following facts

```prolog
bar(a).
bar(b).
```

would be stored as

```text
     bar
     / \
    a   b
   /     \
 bar(a)  bar(b)
```

If we query the database for `bar(a)?`, the search would traverse only
the left branch of the tree. However, the query `bar(X)?` would traverse
and match both branches, returning two results.

The database operations are performed concurrently and the results are
returned in an indeterministic order. The tree search is performed by
recursively spawning goroutines per each branch of the tree until either
failing or finding the match. When the result is found, is is send back
through a Go channel to the querying process.

## Query evaluation and unification

When a query like `same(X, 1)?` is unified with the fact `same(A, A).`
the following is verified:

 1. the names of the atoms are the same,
 2. their *arities* (number of arguments) are the same,
 3. all their arguments can be unified.

When unifying the arguments

* if they are constant, they are compared by type and value,
* wildcard `_` always unifies with anything,
* if variable was not initialized, it with a value by becoming
  equivalent to the value,
* if variable was initialized, its value is unified with the other value.

More information about the unification algorithm can be found in the paper by
Peter Norvig in ["Correcting A Widespread Error in Unification Algorithms"].

The unification of `same(X, 1)` and `same(A, A)` would happen in three steps

 1. `X = A`,
 2. `1 = A`,
 3. `X = 1` by *substituting* `A = 1`, concluding that they unify.

Rules are evaluated by unifying their heads with the query and then
evaluating all the terms in the rule's body.

The evaluation uses *naïve* search by directly traversing all the relevant
branches. No query optimizations are applied.

## Operators

The following operators can be applied to primitive values:
`=`, `!=`, `<`, `<=`, `>`, `>=`, `in`. For example,

```prolog
negative(X) :- X < 0.
```

would be satisfied only for the values of `X` that are negative numbers.

Additionally, the `in` operator can be used to search a substring
within a string, like `hello in "hello, world!"`.

Unlike Prolog, `=` does not perform unification, just checks the value
for equality.

## External data sources

The facts can be read from external sources like standard input
or comma-separated files. To do so, you need to use the `#input`
command:

```prolog
#input foo(source="file.csv", delimiter=",", skip=0, columns="1-2,6")
#input bar(source=stdin, delimiter="\t")
```

It takes the following arguments:

* `source` is either `stdin` or the path to a file,
* field `separator` or `sep` (tab by default),
* how many rows to `skip` before reading the inputs,
* `columns` or `cols` to read - the numbering starts with 1,
  it can be either individual values or ranges `from-to`,
  separated by commas.

The order of the arguments does not matter.

## Grammar

The grammar of Datalo.go is consistent with this [specification],
with minor simplifications and modifications.

```text
program    ::= ( atom ( "." | "~" | "?" ) | rule )* ;
atom       ::= identifier ( "(" term ( "," term )* ")" )? ;
identifier ::= LOWERCASE ( ALPHA | DIGIT | "_" )* ;
term       ::= constant | variable | wildcard ;
string     ::= identifier | "\"" [^"]* "\""
constant   ::= string | number ;
number     ::= ( "+" | "-" )? DIGIT+
variable   ::= UPPERCASE ( ALPHA | DIGIT | "_" )* ;
wildcard   ::= "_" ;
rule       ::= atom ":-" literal ( "," literal )* "." ;
literal    ::= atom | arithmetic ;
arithmetic ::= constant operator constant ;
operator   ::= "=" | "!=" | "<" | "<=" | ">" | ">=" | "in"
```


 [Datalog]: https://en.wikipedia.org/wiki/Datalog
 [specification]: https://datalog-specs.info/vnd_datalog_text/abstract.html
 [unified]: https://en.wikipedia.org/wiki/Unification_(computer_science)
 ["Correcting A Widespread Error in Unification Algorithms"]: https://norvig.com/unify-bug.pdf
