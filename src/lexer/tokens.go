package lexer

import (
	"fmt"
	"github.com/gojinja/gojinja/src/utils/set"
	"golang.org/x/exp/maps"
	"os"
	"regexp"
	"sort"
	"strings"
)

const TokenAdd = "add"
const TokenAssign = "assign"
const TokenColon = "colon"
const TokenComma = "comma"
const TokenDiv = "div"
const TokenDot = "dot"
const TokenEq = "eq"
const TokenFloordiv = "floordiv"
const TokenGt = "gt"
const TokenGteq = "gteq"
const TokenLBrace = "lbrace"
const TokenLBracket = "lbracket"
const TokenLParen = "lparen"
const TokenLt = "lt"
const TokenLteq = "lteq"
const TokenMod = "mod"
const TokenMul = "mul"
const TokenNe = "ne"
const TokenPipe = "pipe"
const TokenPow = "pow"
const TokenRBrace = "rbrace"
const TokenRBracket = "rbracket"
const TokenRParen = "rparen"
const TokenSemicolon = "semicolon"
const TokenSub = "sub"
const TokenTilde = "tilde"
const TokenWhitespace = "whitespace"
const TokenFloat = "float"
const TokenInteger = "integer"
const TokenName = "name"
const TokenString = "string"
const TokenOperator = "operator"
const TokenBlockBegin = "block_begin"
const TokenBlockEnd = "block_end"
const TokenVariableBegin = "variable_begin"
const TokenVariableEnd = "variable_end"
const TokenRawBegin = "raw_begin"
const TokenRawEnd = "raw_end"
const TokenCommentBegin = "comment_begin"
const TokenCommentEnd = "comment_end"
const TokenComment = "comment"
const TokenLinestatementBegin = "linestatement_begin"
const TokenLinestatementEnd = "linestatement_end"
const TokenLinecommentBegin = "linecomment_begin"
const TokenLinecommentEnd = "linecomment_end"
const TokenLinecomment = "linecomment"
const TokenData = "data"
const TokenInitial = "initial"
const TokenEOF = "eof"

var operators = map[string]string{
	"+":  TokenAdd,
	"-":  TokenSub,
	"/":  TokenDiv,
	"//": TokenFloordiv,
	"*":  TokenMul,
	"%":  TokenMod,
	"**": TokenPow,
	"~":  TokenTilde,
	"[":  TokenLBracket,
	"]":  TokenRBracket,
	"(":  TokenLParen,
	")":  TokenRParen,
	"{":  TokenLBrace,
	"}":  TokenRBrace,
	"==": TokenEq,
	"!=": TokenNe,
	">":  TokenGt,
	">=": TokenGteq,
	"<":  TokenLt,
	"<=": TokenLteq,
	"=":  TokenAssign,
	".":  TokenDot,
	":":  TokenColon,
	"|":  TokenPipe,
	",":  TokenComma,
	";":  TokenSemicolon,
}

var reverseOperators = rev(operators)

func rev(ops map[string]string) map[string]string {
	rev := make(map[string]string)
	for k, v := range ops {
		rev[v] = k
	}
	if len(rev) != len(ops) {
		_, _ = fmt.Fprintln(os.Stderr, "operators dropped")
		os.Exit(1)
	}
	return rev
}

var operatorRe = getOperatorRe(operators)

func getOperatorRe(ops map[string]string) *regexp.Regexp {
	els := maps.Keys(ops)
	sort.Slice(els, func(i int, j int) bool {
		return len(els[i]) > len(els[j])
	})
	for i, el := range els {
		els[i] = regexp.QuoteMeta(el)
	}
	pat := strings.Join(els, "|")
	return regexp.MustCompile("^" + pat)
}

var ignoredTokens = set.FrozenFromElems(
	TokenCommentBegin,
	TokenComment,
	TokenCommentEnd,
	TokenWhitespace,
	TokenLinecommentBegin,
	TokenLinecommentEnd,
	TokenLinecomment,
)

var ignoreIfEmpty = set.FrozenFromElems(
	TokenWhitespace,
	TokenData,
	TokenComment,
	TokenLinecomment,
)

func describeTokenType(tokenType string) string {
	if op, ok := reverseOperators[tokenType]; ok {
		return op
	}

	switch tokenType {
	case TokenCommentBegin:
		return "begin of comment"
	case TokenCommentEnd:
		return "end of comment"
	case TokenComment:
		return "comment"
	case TokenLinecomment:
		return "comment"
	case TokenBlockBegin:
		return "begin of statement block"
	case TokenBlockEnd:
		return "end of statement block"
	case TokenVariableBegin:
		return "begin of print statement"
	case TokenVariableEnd:
		return "end of print statement"
	case TokenLinestatementBegin:
		return "begin of line statement"
	case TokenLinestatementEnd:
		return "end of line statement"
	case TokenData:
		return "template data / text"
	case TokenEOF:
		return "end of template"
	default:
		return tokenType
	}
}

func DescribeToken(token Token) string {
	if token.Type == TokenName {
		return fmt.Sprint(token.Value)
	}

	return describeTokenType(token.Type)
}

func DescribeTokenExpr(expr string) string {
	if strings.Contains(expr, ":") {
		res := strings.SplitN(expr, ":", 1)
		if res[0] == TokenName {
			return res[1]
		}
		return describeTokenType(res[0])
	}
	return describeTokenType(expr)
}

type Token struct {
	Lineno int
	Type   string
	Value  any
}

func (t Token) String() string {
	return DescribeToken(t)
}

func (t Token) Test(expr string) bool {
	if t.Type == expr {
		return true
	}
	if strings.Contains(expr, ":") {
		res := strings.SplitN(expr, ":", 2)
		return res[0] == t.Type && res[1] == t.Value
	}
	return false
}

func (t Token) TestAny(exprs ...string) bool {
	for _, expr := range exprs {
		if t.Test(expr) {
			return true
		}
	}
	return false
}
