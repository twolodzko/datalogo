package parser

import (
	"fmt"
	"os/user"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/twolodzko/datalogo/datalog"
)

// Examples:
//
//	#input foo(source="file.csv", delimiter=",", skip=0, columns="1:5")
//	#input bar(source=stdin, delimiter="\t", columns="1-2,6")
type Input struct {
	Name      string
	Source    string
	Separator string
	Skip      int
	Columns   []int
}

func (inp Input) ParseLine(line string) (datalog.Atom, error) {
	atom := datalog.Atom{
		Name: inp.Name,
	}
	fields := strings.Split(line, inp.Separator)

	if len(inp.Columns) > 0 {
		for _, i := range inp.Columns {
			if i >= len(fields) {
				return datalog.Atom{}, fmt.Errorf("missing column: %d", i+1)
			}
			field := strings.TrimSpace(fields[i])
			term, err := parseTerm(field)
			if err != nil {
				return datalog.Atom{}, err
			}
			atom.Args = append(atom.Args, term)
		}
	} else {
		for _, field := range fields {
			field = strings.TrimSpace(field)
			term, err := parseTerm(field)
			if err != nil {
				return datalog.Atom{}, err
			}
			atom.Args = append(atom.Args, term)
		}
	}

	return atom, nil
}

func (p *Parser) readInput() (Input, error) {
	var res Input

	name, err := p.readToken()
	if err != nil {
		return Input{}, err
	}
	if !isIdentifier(name) {
		return Input{}, UnexpectedToken{name}
	}
	res.Name = name

	// (
	if err := p.expect("("); err != nil {
		return Input{}, err
	}

	for {
		// key
		key, err := p.readToken()
		if err != nil {
			return Input{}, err
		}

		// =
		if err := p.expect("="); err != nil {
			return Input{}, err
		}

		// val
		val, err := p.readTerm()
		if err != nil {
			return Input{}, err
		}

		switch key {
		case "source":
			switch val := val.(type) {
			case datalog.String:
				if val == "stdin" {
					res.Source = string(val)
				} else {
					res.Source, err = parsePath(string(val))
					if err != nil {
						return Input{}, err
					}
				}
			default:
				return Input{}, WrongValue{key, val}
			}
		case "separator", "sep":
			switch val := val.(type) {
			case datalog.String:
				res.Separator = string(val)
			default:
				return Input{}, WrongValue{key, val}
			}
		case "skip":
			switch val := val.(type) {
			case int:
				res.Skip = val
			default:
				return Input{}, WrongValue{key, val}
			}
		case "columns", "cols":
			switch val := val.(type) {
			case datalog.String:
				cols, err := parseColumns(string(val))
				if err != nil {
					return Input{}, err
				}
				res.Columns = cols
			default:
				return Input{}, WrongValue{key, val}
			}
		default:
			return Input{}, fmt.Errorf("unknown key: %s", key)
		}

		token, err := p.readToken()
		if err != nil {
			return Input{}, err
		}
		if token == ")" {
			break
		} else if token != "," {
			return Input{}, fmt.Errorf("unknown key: %s", key)
		}
	}

	if res.Separator == "" {
		switch {
		case strings.HasSuffix(res.Source, ".csv"):
			res.Separator = ","
		default:
			res.Separator = "\t"
		}
	}
	if res.Source == "" {
		res.Source = "stdin"
	}

	return res, nil
}

func parseColumns(input string) ([]int, error) {
	var out []int
	fields := strings.Split(input, ",")
	for _, field := range fields {
		vals := strings.Split(field, "-")
		switch len(vals) {
		case 1:
			val, err := strconv.Atoi(vals[0])
			if err != nil {
				return nil, err
			}
			out = append(out, val-1)
		case 2:
			lower, err := strconv.Atoi(vals[0])
			if err != nil {
				return nil, err
			}
			upper, err := strconv.Atoi(vals[0])
			if err != nil {
				return nil, err
			}
			for i := lower - 1; i <= upper+1; i++ {
				out = append(out, i)
			}
		default:
			return nil, fmt.Errorf("invalid range selector: %s", field)
		}
	}
	sort.Ints(out)
	out = slices.Compact(out)
	return out, nil
}

func parsePath(path string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	dir := usr.HomeDir
	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(dir, path[2:])
	}
	return filepath.Abs(path)
}

type WrongValue struct {
	key string
	val any
}

func (e WrongValue) Error() string {
	return fmt.Sprintf("argument of %s = %v has an invalid value", e.key, e.val)
}
