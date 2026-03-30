package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

    "github.com/strickyak/stipulate/evaluator"
)

var Defs = make(map[string]uint16)

var Comment = regexp.MustCompile(`^\s*([*]|[;]|$)`).FindStringSubmatch

var ParseOp = regexp.MustCompile(`^([A-Za-z0-9_.$@]+[:]?)?(\s+([A-Za-z0-9_.]+)(\s+([^ ;]+))?)?(.*)`).FindStringSubmatch

// TakeNone are the niladic opcodes.
var TakeNone = map[string]bool{
	"nop":  true,
	"rts":  true,
	"rti":  true,
	"clra": true,
	"clrb": true,
	"coma": true,
	"comb": true,
	"inca": true,
	"incb": true,
	"deca": true,
	"decb": true,
	"swi":  true,
	"swi2": true,
	"swi3": true,
	"daa":  true,
	"cwai": true,
	"sync": true,
	"emod": true,
	"else": true,
	"endc": true,
}

// TakeRest may have spaces in the argument.
var TakeRest = map[string]bool{
	"nam": true,
	"ttl": true,
	"fcc": true,
	"fcs": true,
}

var Cond = map[string]func(string) bool{
	"if": func(a string) bool {
		b, _ := Defs[a]
        b = evaluator.Evaluate(a, Defs)
		return (b != 0)
	},
	"ifeq": func(a string) bool {
		b, _ := Defs[a]
        b = evaluator.Evaluate(a, Defs)
		return (b == 0)
	},
	"ifne": func(a string) bool {
		b, _ := Defs[a]
        b = evaluator.Evaluate(a, Defs)
		return (b != 0)
	},
	"ifgt": func(a string) bool {
		b, _ := Defs[a]
        b = evaluator.Evaluate(a, Defs)
		return (int16(b) > 0)  // switch to signed comparison
	},
	"ifge": func(a string) bool {
		b, _ := Defs[a]
        b = evaluator.Evaluate(a, Defs)
		return (int16(b) >= 0)  // switch to signed comparison
	},
	"iflt": func(a string) bool {
		b, _ := Defs[a]
        b = evaluator.Evaluate(a, Defs)
		return (int16(b) < 0)  // switch to signed comparison
	},
	"ifle": func(a string) bool {
		b, _ := Defs[a]
        b = evaluator.Evaluate(a, Defs)
		return (int16(b) <= 0)  // switch to signed comparison
	},
	"ifp1": func(a string) bool {
		return true
	},
	"ifdef": func(a string) bool {
		_, ok := Defs[a]
		return ok
	},
	"ifnd": func(a string) bool {
		_, ok := Defs[a]
		return !ok
	},
}

// DefineFlag acts as a bridge between Go's flag package and our definitions map.
type DefineFlag struct {
	Defs map[string]uint16
}

// String is required by the flag.Value interface.
// It represents the default value formatted as a string.
func (d *DefineFlag) String() string {
	return ""
}

// Set is called by the flag package every time it encounters our flag (e.g., "-D").
// It implements the flag.Value interface.
func (d *DefineFlag) Set(value string) error {
	evaluator.Define(value, d.Defs)
	return nil
}

func main() {
	defFlag := &DefineFlag{Defs: Defs}
	// The flag package will now call defFlag.Set() every time it sees "-D"
	flag.Var(defFlag, "D", "Define a symbol (Format: SYM or SYM=VAL)")
	flag.Parse()

	for _, a := range flag.Args() {
		Defs[a] = 1
	}

	r := os.Stdin
	scanner := bufio.NewScanner(r)
	i := 0
	active := true
	for scanner.Scan() {
		i++
		line := scanner.Text()
		line = strings.TrimRight(line, " \t\r\n")
		log.Printf("%v", line)

		comment := Comment(line)
		if comment != nil {
			log.Printf("====== (comment)")
            if line == "" {
                fmt.Printf("\n")
            } else {
                fmt.Printf("----------------------  %s\n", line)
            }
			continue
		}

		m := ParseOp(line)
		if m == nil {
            NoIndent(Format("WHAT? %s", line))
			continue
		}

		log.Printf("====== %q || %q || %q || %q", m[1], m[2], m[3], m[4])
		label, op, arg, remark := m[1], m[3], m[5], m[6]
		_ = label
		_ = remark
		_, tn := TakeNone[op]
		_, tr := TakeRest[op]
		_ = tn
		_ = tr

		c, _ := Cond[op]
		switch {
		case c != nil:
			Push(active)
			active = c(arg)
			HalfIndent("{{{ ", line)

		case op == "else":
			active = !active
			HalfIndent("}={ ", line)

		case op == "endc":
			HalfIndent("}}} ", line)
			active = Pop()

		default:
			if AllActive(active) {
				NoIndent(line)
			} else {
				FullIndent(line)
			}

		}

	}
}

var CondStack []bool

func Push(a bool) {
	CondStack = append(CondStack, a)
}
func Pop() bool {
	last := len(CondStack) - 1
	z := CondStack[last]
	CondStack = CondStack[:last]
	return z
}
func AllActive(active bool) bool {
	for _, b := range CondStack {
		if !b {
			return false
		}
	}
	return active
}
func NoIndent(line string) {
	fmt.Printf("%s\n", line)
}
func HalfIndent(decoration, line string) {
	for _ = range CondStack {
		fmt.Printf("%s", decoration)
	}
	fmt.Printf("%s\n", line)
}
func FullIndent(line string) {
	fmt.Printf(";;;;;;;;;;;;;;;;;;;;;;;;;;;;;;  %s\n", line)
}
var Format = fmt.Sprintf
