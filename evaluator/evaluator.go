package evaluator // by Gemini 3.1 Pro

import (
	"strconv"
	"strings"
	"unicode"
)

// Evaluate evaluates an lwasm-style arithmetic expression.
// All calculations natively wrap modulo 2^16 due to the uint16 type.
func Evaluate(expr string, defs map[string]uint16) uint16 {
	p := &parser{
		expr: []rune(expr),
		defs: defs,
	}
	p.advance() // Prime the first token
	return p.parseExpr(0)
}

type token struct {
	kind  string
	value string
}

type parser struct {
	expr []rune
	pos  int
	defs map[string]uint16
	cur  token
}

// precedences maps operators to their standard C-like precedence level.
var precedences = map[string]int{
	"|":  1,
	"^":  2,
	"&":  3,
	"<<": 4, ">>": 4,
	"+":  5, "-":  5,
	"*":  6, "/":  6, "%":  6,
}

// parseExpr processes binary expressions using Precedence Climbing.
func (p *parser) parseExpr(minPrec int) uint16 {
	val := p.parseUnary()

	for {
		prec, ok := precedences[p.cur.kind]
		if !ok || prec < minPrec {
			break
		}
		
		op := p.cur.kind
		p.advance()

		// Evaluate the right-hand side using a higher precedence threshold.
		rhs := p.parseExpr(prec + 1)

		switch op {
		case "+":  val += rhs
		case "-":  val -= rhs
		case "*":  val *= rhs
		case "/":
			if rhs != 0 { val /= rhs } else { val = 0 } // Prevent division by zero panic
		case "%":
			if rhs != 0 { val %= rhs } else { val = 0 }
		case "&":  val &= rhs
		case "|":  val |= rhs
		case "^":  val ^= rhs
		case "<<": val <<= rhs
		case ">>": val >>= rhs
		}
	}
	return val
}

// parseUnary handles positive, negative, and bitwise NOT prefixes.
func (p *parser) parseUnary() uint16 {
	if p.cur.kind == "+" {
		p.advance()
		return p.parseUnary()
	}
	if p.cur.kind == "-" {
		p.advance()
		return -p.parseUnary() // Two's complement negation is handled natively by uint16
	}
	if p.cur.kind == "~" {
		p.advance()
		return ^p.parseUnary()
	}
	return p.parsePrimary()
}

// parsePrimary evaluates base values: numbers, symbols, and nested expressions.
func (p *parser) parsePrimary() uint16 {
	var val uint16
	if p.cur.kind == "num" {
		val = parseNum(p.cur.value)
		p.advance()
	} else if p.cur.kind == "sym" {
		val = p.defs[p.cur.value] // Naturally evaluates to 0 if the symbol is undefined
		p.advance()
	} else if p.cur.kind == "(" {
		p.advance()
		val = p.parseExpr(0)
		if p.cur.kind == ")" {
			p.advance()
		}
	}
	return val
}

// advance acts as the lexer, moving to the next valid token.
func (p *parser) advance() {
	// Skip whitespace
	for p.pos < len(p.expr) && unicode.IsSpace(p.expr[p.pos]) {
		p.pos++
	}
	if p.pos >= len(p.expr) {
		p.cur = token{"EOF", ""}
		return
	}
	
	c := p.expr[p.pos]

	// Single-character operators
	switch c {
	case '+', '-', '*', '/', '%', '(', ')', '&', '|', '^', '~':
		p.pos++
		p.cur = token{string(c), string(c)}
		return
	case '<', '>': // Left and Right Shifts
		if p.pos+1 < len(p.expr) && p.expr[p.pos+1] == c {
			p.pos += 2
			p.cur = token{string(c) + string(c), string(c) + string(c)}
			return
		}
		p.pos++
		p.cur = token{string(c), string(c)}
		return
	case '$': // Base 16 constants
		p.pos++
		start := p.pos
		for p.pos < len(p.expr) && isHexDigit(p.expr[p.pos]) {
			p.pos++
		}
		p.cur = token{"num", "$" + string(p.expr[start:p.pos])}
		return
	}

	// Base 10 constants
	if unicode.IsDigit(c) {
		start := p.pos
		for p.pos < len(p.expr) && unicode.IsDigit(p.expr[p.pos]) {
			p.pos++
		}
		p.cur = token{"num", string(p.expr[start:p.pos])}
		return
	}

	// Symbols (variables)
	if isSymStart(c) {
		start := p.pos
		for p.pos < len(p.expr) && isSymChar(p.expr[p.pos]) {
			p.pos++
		}
		p.cur = token{"sym", string(p.expr[start:p.pos])}
		return
	}

	// Fallback/Catch-all
	p.pos++
	p.cur = token{string(c), string(c)}
}

// Helper methods to identify valid characters
func isSymStart(c rune) bool {
	return unicode.IsLetter(c) || c == '_' || c == '.' || c == '@' || c == '?'
}

func isSymChar(c rune) bool {
	// ADDED: c == '$' to allow dollar signs inside the symbol
	return isSymStart(c) || unicode.IsDigit(c) || c == '$'
}

func isHexDigit(c rune) bool {
	return unicode.IsDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

// parseNum converts the parsed string into a uint16, respecting base 10 and 16 rules.
func parseNum(s string) uint16 {
	if strings.HasPrefix(s, "$") {
		v, _ := strconv.ParseUint(s[1:], 16, 16)
		return uint16(v)
	}
	v, _ := strconv.ParseUint(s, 10, 16)
	return uint16(v)
}
