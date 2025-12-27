package lexer

import (
	"fmt"
	"regexp"
	"slices"
)

type lexer struct {
	patterns []pattern
	tokens   []Token
	src      string
	pos      int
}

func (lex *lexer) atEof() bool {
	return lex.pos >= len(lex.src)
}

func (lex *lexer) remainder() string {
	return lex.src[lex.pos:]
}

func (lex *lexer) push(token Token) {
	lex.tokens = append(lex.tokens, token)
}

func (lex *lexer) advance(n int) {
	lex.pos += n
}

type pattern struct {
	regex   *regexp.Regexp
	handler patternHandler
}

type patternHandler func(lex *lexer, regex *regexp.Regexp)

func Tokenize(src string) []Token {
	lex := createLexer(src)
	matchHandler := getMatchHandler(lex)

	for !lex.atEof() {
		matched := slices.ContainsFunc(lex.patterns, matchHandler)

		if !matched {
			panic(fmt.Sprintf("Unrecognized token near %s\n", lex.remainder()))
		}
	}

	lex.push(Token{EOF, ""})
	return lex.tokens
}

func getMatchHandler(lex *lexer) func(p pattern) bool {
	return func(p pattern) bool {
		loc := p.regex.FindStringIndex(lex.remainder())

		if loc == nil || loc[0] != 0 {
			return false
		}

		p.handler(lex, p.regex)
		return true
	}
}

func defaultHandler(type_ TokenType) patternHandler {
	return func(lex *lexer, regex *regexp.Regexp) {
		match := regex.FindString(lex.remainder())
		lex.push(Token{type_, match})
		lex.advance(len(match))
	}
}

func skipHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindStringIndex(lex.remainder())
	lex.advance(match[1])
}

func symbolHandler(lex *lexer, regex *regexp.Regexp) {
	match := regex.FindString(lex.remainder())

	type_, ok := ReservedKeywords[match]

	if !ok { type_ = Identifier }

	lex.push(Token{type_, match})
	lex.advance(len(match))
}

func createLexer(src string) *lexer {
	return &lexer{
		pos:    0,
		src:    src,
		tokens: make([]Token, 0),
		patterns: []pattern{
			{regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`), symbolHandler},
			{regexp.MustCompile(`[0-9]+(\.[0-9]+)?`), defaultHandler(Number)},
			{regexp.MustCompile(`'[^']'`), defaultHandler(Char)},
			{regexp.MustCompile(`\[`), defaultHandler(LBracket)},
			{regexp.MustCompile(`\]`), defaultHandler(RBracket)},
			{regexp.MustCompile(`\(`), defaultHandler(LParen)},
			{regexp.MustCompile(`\)`), defaultHandler(RParen)},
			{regexp.MustCompile(`\{`), defaultHandler(LCurly)},
			{regexp.MustCompile(`\}`), defaultHandler(RCurly)},
			{regexp.MustCompile(`->`), defaultHandler(Lambda)},
			{regexp.MustCompile(`:`), defaultHandler(Colon)},
			{regexp.MustCompile(`,`), defaultHandler(Comma)},
			{regexp.MustCompile(`\.\.`), defaultHandler(PipeDot)},
			{regexp.MustCompile(`>>`), defaultHandler(PipeL)},
			{regexp.MustCompile(`<<`), defaultHandler(PipeR)},
			{regexp.MustCompile(`=`), defaultHandler(Equals)},
			{regexp.MustCompile(`!=`), defaultHandler(NotEquals)},
			{regexp.MustCompile(`<`), defaultHandler(Less)},
			{regexp.MustCompile(`<=`), defaultHandler(LessEquals)},
			{regexp.MustCompile(`>`), defaultHandler(Greater)},
			{regexp.MustCompile(`>=`), defaultHandler(GreaterEquals)},
			{regexp.MustCompile(`\+`), defaultHandler(Plus)},
			{regexp.MustCompile(`-`), defaultHandler(Dash)},
			{regexp.MustCompile(`/`), defaultHandler(Slash)},
			{regexp.MustCompile(`\*`), defaultHandler(Star)},
			{regexp.MustCompile(`%`), defaultHandler(Percent)},
			{regexp.MustCompile(`~~[^~]*~~`), skipHandler},
			{regexp.MustCompile(`~.*`), skipHandler},
			{regexp.MustCompile(`\s+`), skipHandler},
		},
	}
}