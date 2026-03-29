package evaluator

import (
	"strconv"
	"strings"
)

// Define processes command-line symbol definitions and populates the defs map.
func Define(s string, defs map[string]uint16) {
	// Trim any accidental whitespace from the command line argument
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}

	eqIdx := strings.IndexByte(s, '=')

	// Format 1: No '=' sign, symbol is defined as 1
	if eqIdx == -1 {
		defs[s] = 1
		return
	}

	// Format 2: SYMBOL=VALUE
	sym := strings.TrimSpace(s[:eqIdx])
	valStr := strings.TrimSpace(s[eqIdx+1:])

	var val uint64
	var err error

	// Determine base and parse accordingly
	if strings.HasPrefix(valStr, "$") {
		// Hexadecimal with '$' prefix
		val, err = strconv.ParseUint(valStr[1:], 16, 16)
	} else if strings.HasPrefix(strings.ToLower(valStr), "0x") {
		// Hexadecimal with '0x' or '0X' prefix
		val, err = strconv.ParseUint(valStr[2:], 16, 16)
	} else {
		// Decimal
		val, err = strconv.ParseUint(valStr, 10, 16)
	}

	// If a user passes a malformed number (e.g., "SYM=ABC"), err won't be nil.
	// ParseUint naturally returns 0 on error, which safely defaults the symbol to 0.
	if err == nil {
		defs[sym] = uint16(val)
	} else {
		defs[sym] = 0 // Or you can choose to skip assignment on bad input: return
	}
}
