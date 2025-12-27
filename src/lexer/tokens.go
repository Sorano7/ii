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

	// -- Brackets
	LParen
	RParen
	LBracket
	RBracket
	LCurly
	RCurly

	// -- Operators
	Colon
	Lambda
	PipeL
	PipeR
	PipeDot
	Comma

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

var TokenTypeNames = map[TokenType]string{
	EOF:           "EOF",
	Number:        "Number",
	Char:          "Char",
	LParen:        "LParen",
	RParen:        "RParen",
	LBracket:      "LBracket",
	RBracket:      "RBracket",
	LCurly:        "LCurly",
	RCurly:        "RCurly",
	Colon:         "Colon",
	Lambda:        "Lambda",
	PipeL:         "PipeL",
	PipeR:         "PipeR",
	PipeDot:       "PipeDot",
	Comma:         "Comma",
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