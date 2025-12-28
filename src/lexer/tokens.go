package lexer

import (
	"fmt"
	"slices"
)

type TokenType int

const (
	EOF TokenType = iota

	// -- Literals
	Number
	Char
	String

	// -- Brackets
	LParen
	RParen
	LBracket
	RBracket
	LCurly
	RCurly

	// -- Operators
	Assignment
	Lambda
	RPipe
	LPipe
	Diamond
	Dot
	Comma
	If
	Else
	Ellipsis

	// -- Comparison
	Equals
	NotEquals
	Less
	LessEquals
	Greater
	GreaterEquals

	// -- Arithmetic
	Plus
	Dash
	Slash
	Star
	Percent

	// -- Keywords
	Identifier
)

var ReservedKeywords = map[string]TokenType{

}

var ComparisonOps = []TokenType {
	Equals, NotEquals, Less, LessEquals, Greater, GreaterEquals,
}

var ArithmeticOps = []TokenType {
	Plus, Dash, Slash, Star, Percent,
}

var TokenTypeNames = map[TokenType]string{
	EOF:           "EOF",
	Number:        "Number",
	Char:          "Char",
	String:        "String",
	LParen:        "LParen",
	RParen:        "RParen",
	LBracket:      "LBracket",
	RBracket:      "RBracket",
	LCurly:        "LCurly",
	RCurly:        "RCurly",
	Assignment:    "Assignment",
	Lambda:        "Lambda",
	LPipe:         "LPipe",
	RPipe:         "RPipe",
	Diamond:       "Diamond",
	Dot:           "Dot",
	Comma:         "Comma",
	If:            "If",
	Else:          "Else",
	Ellipsis:      "Ellipsis",
	Equals:        "Equals",
	NotEquals:     "NotEquals",
	Less:          "Less",
	LessEquals:    "LessEquals",
	Greater:       "Greater",
	GreaterEquals: "GreaterEquals",
	Plus:          "Plus",
	Dash:          "Dash",
	Slash:         "Slash",
	Star:          "Star",
	Percent:       "Percent",
	Identifier:    "Identifier",
}

func (tp TokenType) String() string {
	return TokenTypeNames[tp]
}

type Token struct {
	Type  TokenType
	Value string
}

func (t Token) In(expected ...TokenType) bool {
	return slices.Contains(expected, t.Type)
}

func (t Token) String() string {
	value := ""

	if t.In(Identifier, Number) {
		value = t.Value
	}

	return fmt.Sprintf("%s (%s)", t.Type, value)
}