package evaluator // by Gemini 3.1 Pro

import "testing"

func TestEvaluate(t *testing.T) {
	// A shared set of definitions for our tests
	defs := map[string]uint16{
		"TEN":     10,
		"TWENTY":  20,
		"MAX_VAL": 65535,
		"VAR$1":   50,
		"A$B$C":   5,
	}

	tests := []struct {
		name     string
		expr     string
		expected uint16
	}{
		// 1. Base Constants
		{"Base 10 constant", "123", 123},
		{"Base 16 constant (lower)", "$1a", 26},
		{"Base 16 constant (upper)", "$1A", 26},
		{"Base 16 constant (max)", "$FFFF", 65535},

		// 2. Symbols
		{"Known symbol", "TEN", 10},
		{"Unknown symbol defaults to zero", "UNDEFINED", 0},
		{"Symbol with dollar sign", "VAR$1", 50},
		{"Multiple dollar signs in symbol", "A$B$C", 5},

		// 3. Basic Arithmetic
		{"Addition", "TEN + TWENTY", 30},
		{"Subtraction", "TWENTY - TEN", 10},
		{"Multiplication", "TEN * 3", 30},
		{"Division", "TWENTY / 2", 10},
		{"Modulo", "TWENTY % 3", 2},

		// 4. Precedence & Grouping
		{"Operator precedence (* over +)", "2 + 3 * 4", 14},
		{"Operator precedence (/ over -)", "20 - 10 / 2", 15},
		{"Parentheses precedence", "(2 + 3) * 4", 20},
		{"Nested parentheses", "((2 + 3) * 2) - 1", 9},

		// 5. Unary Operators
		{"Unary positive", "+5", 5},
		{"Unary negative (underflow)", "-1", 65535},
		{"Unary negative on symbol", "-TEN", 65526},

		// 6. Bitwise Operations
		{"Bitwise AND", "3 & 5", 1},
		{"Bitwise OR", "1 | 2", 3},
		{"Bitwise XOR", "1 ^ 3", 2},
		{"Bitwise NOT", "~0", 65535},
		{"Left shift", "1 << 4", 16},
		{"Right shift", "16 >> 2", 4},

		// 7. Modulo 2^16 Arithmetic Constraints
		{"Modulo 2^16 overflow", "65535 + 2", 1},
		{"Modulo 2^16 underflow", "0 - 1", 65535},
		{"Multiplication overflow", "32768 * 2", 0},

		// 8. Graceful Fails / Zero Panics
		{"Division by zero", "TEN / 0", 0},
		{"Modulo by zero", "TEN % 0", 0},

		// 9. Complex Expressions
		{"Complex lwasm string", "VAR$1 * (TEN + $A) / 2", 500}, // 50 * (10 + 10) / 2
		{"Whitespaces handled", "  TEN   +  \t $5 ", 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Evaluate(tt.expr, defs)
			if result != tt.expected {
				t.Errorf("Evaluate(%q) = %d; want %d", tt.expr, result, tt.expected)
			}
		})
	}
}
