package parser

import (
	"fmt"
	"github.com/gojinja/gojinja/src/errors"
	"github.com/gojinja/gojinja/src/extensions"
	"github.com/gojinja/gojinja/src/lexer"
	"github.com/gojinja/gojinja/src/nodes"
	"github.com/gojinja/gojinja/src/utils/set"
	"github.com/gojinja/gojinja/src/utils/slices"
	"github.com/gojinja/gojinja/src/utils/stack"
	"strings"
)

var statementKeywords = set.FrozenFromElems(
	"for",
	"if",
	"block",
	"extends",
	"print",
	"macro",
	"include",
	"from",
	"import",
	"set",
	"with",
	"autoescape",
)

var compareOperators = set.FrozenFromElems(
	"eq", "ne", "lt", "lteq", "gt", "gteq",
)

func makeBinaryOpNode(left, right nodes.Node, op string, lineno int) nodes.Node {
	return &nodes.BinOp{
		Left:       left,
		Right:      right,
		Op:         op,
		NodeCommon: nodes.NodeCommon{Lineno: lineno},
	}
}

type extensionParser = func(p extensions.IParser) ([]nodes.Node, error)

type parser struct {
	stream                *lexer.TokenStream
	name, filename, state *string
	closed                bool
	extensions            map[string]extensionParser
	lastIdentifier        int
	tagStack              *stack.Stack[string]
	endTokenStack         *stack.Stack[[]string]
}

var _ extensions.IParser = &parser{}

func NewParser(stream *lexer.TokenStream, extensions []extensions.IExtension, name, filename, state *string) *parser {
	taggedExtensions := make(map[string]extensionParser, 0)
	for _, extension := range extensions {
		for _, tag := range extension.Tags() {
			taggedExtensions[tag] = extension.Parse
		}
	}

	return &parser{
		stream:         stream,
		name:           name,
		filename:       filename,
		state:          state,
		closed:         false,
		extensions:     taggedExtensions,
		lastIdentifier: 0,
		tagStack:       stack.New[string](),
		endTokenStack:  stack.New[[]string](),
	}
}

// Parse parses the whole template into a `Template` node.
func (p *parser) Parse() (*nodes.Template, error) {
	body, err := p.subparse(nil)
	if err != nil {
		return nil, err
	}

	// TODO set environment
	return &nodes.Template{
		Body:       body,
		NodeCommon: nodes.NodeCommon{Lineno: 1},
	}, nil
}

func (p *parser) subparse(endTokens []string) ([]nodes.Node, error) {
	body := make([]nodes.Node, 0)
	dataBuffer := make([]nodes.Node, 0)
	addData := func(node nodes.Node) {
		dataBuffer = append(dataBuffer, node)
	}

	if endTokens != nil {
		p.endTokenStack.Push(endTokens)
		defer p.endTokenStack.Pop()
	}

	flushData := func() {
		if len(dataBuffer) > 0 {
			lineno := dataBuffer[0].GetLineno()
			body = append(body, &nodes.Output{
				Nodes:      dataBuffer,
				NodeCommon: nodes.NodeCommon{Lineno: lineno},
			})
			dataBuffer = make([]nodes.Node, 0)
		}
	}

	for p.stream.Bool() {
		token := p.stream.Current()
		switch token.Type {
		case lexer.TokenData:
			if token.Value != "" {
				// type assert is safe, because token.Type == lexer.TokenData
				addData(&nodes.TemplateData{
					Data:       token.Value.(string),
					NodeCommon: nodes.NodeCommon{Lineno: token.Lineno},
				})
			}
			p.stream.Next()
		case lexer.TokenVariableBegin:
			p.stream.Next()
			tuple, err := p.parseTuple(false, true, nil, false)
			if err != nil {
				return nil, err
			}
			addData(tuple)
			if _, err := p.stream.Expect(lexer.TokenVariableEnd); err != nil {
				return nil, err
			}
		case lexer.TokenBlockBegin:
			flushData()
			p.stream.Next()
			if endTokens != nil && p.stream.Current().TestAny(endTokens...) {
				return body, nil
			}
			rvs, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			body = append(body, rvs...)
			if _, err := p.stream.Expect(lexer.TokenBlockEnd); err != nil {
				return nil, err
			}
		default:
			return nil, p.fail("internal parsing error", nil)
		}
	}

	flushData()
	return body, nil
}

func (p *parser) parseTuple(simplified bool, withCondexpr bool, extraEndRules []string, explicitParentheses bool) (nodes.Node, error) {
	lineno := p.stream.Current().Lineno
	var parse func() (nodes.Node, error)
	if simplified {
		parse = p.parsePrimary
	} else {
		parse = func() (nodes.Node, error) {
			return p.parseExpression(withCondexpr)
		}
	}

	var args []nodes.Node
	var isTuple bool

	for {
		if len(args) > 0 {
			if _, err := p.stream.Expect(lexer.TokenComma); err != nil {
				return nil, err
			}
		}
		if p.isTupleEnd(extraEndRules) {
			break
		}

		arg, err := parse()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		if p.stream.Current().Type == lexer.TokenComma {
			isTuple = true
		} else {
			break
		}
		lineno = p.stream.Current().Lineno
	}

	if !isTuple {
		if len(args) > 0 {
			return args[0], nil
		}
		// if we don't have explicit parentheses, an empty tuple is
		// not a valid expression.  This would mean nothing (literally
		// nothing) in the spot of an expression would be an empty
		// tuple.
		if !explicitParentheses {
			return nil, p.fail(fmt.Sprintf("Expected an expression, got %s", p.stream.Current()), nil)
		}
	}

	return &nodes.Tuple{
		Items:      args,
		Ctx:        "load",
		NodeCommon: nodes.NodeCommon{Lineno: lineno},
	}, nil
}

func (p *parser) parsePrimary() (nodes.Node, error) {
	token := p.stream.Current()
	var node nodes.Node

	switch token.Type {
	case lexer.TokenName:
		switch token.Value {
		case "True", "False", "true", "false":
			node = &nodes.Const{
				Value:      token.Value == "true" || token.Value == "True",
				NodeCommon: nodes.NodeCommon{Lineno: token.Lineno},
			}
		case "None", "none":
			node = &nodes.Const{
				Value:      nil,
				NodeCommon: nodes.NodeCommon{Lineno: token.Lineno},
			}
		default:
			node = &nodes.Name{
				Name:       token.Value.(string),
				Ctx:        "load",
				NodeCommon: nodes.NodeCommon{Lineno: token.Lineno},
			}
		}
		p.stream.Next()
		return node, nil
	case lexer.TokenString:
		p.stream.Next()
		buf := []string{token.Value.(string)}
		lineno := token.Lineno
		for p.stream.Current().Type == lexer.TokenString {
			buf = append(buf, p.stream.Current().Value.(string))
			p.stream.Next()
		}
		return &nodes.Const{
			Value:      strings.Join(buf, ""),
			NodeCommon: nodes.NodeCommon{Lineno: lineno},
		}, nil
	case lexer.TokenInteger, lexer.TokenFloat:
		p.stream.Next()
		return &nodes.Const{
			Value:      token.Value,
			NodeCommon: nodes.NodeCommon{Lineno: token.Lineno},
		}, nil
	case lexer.TokenLParen:
		p.stream.Next()
		node, err := p.parseTuple(false, true, nil, true)
		if err != nil {
			return nil, err
		}
		if _, err := p.stream.Expect(lexer.TokenRParen); err != nil {
			return nil, err
		}
		return node, nil
	case lexer.TokenLBracket:
		return p.parseList()
	case lexer.TokenLBrace:
		return p.parseDict()
	default:
		return nil, p.fail(fmt.Sprintf("unexpected %q", lexer.DescribeToken(token)), &token.Lineno)
	}
}

func (p *parser) parseExpression(withCondexpr bool) (nodes.Node, error) {
	if withCondexpr {
		return p.parseCondexpr()
	}
	return p.parseOr()
}

func (p *parser) parseCondexpr() (nodes.Node, error) {
	lineno := p.stream.Current().Lineno
	expr1, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	var expr3 nodes.Node

	for p.stream.SkipIf("name:if") {
		expr2, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.stream.SkipIf("name:else") {
			expr3, err = p.parseCondexpr()
			if err != nil {
				return nil, err
			}
		} else {
			expr3 = nil
		}
		expr1 = &nodes.CondExpr{
			Test:       expr2,
			Expr1:      expr1,
			Expr2:      expr3,
			NodeCommon: nodes.NodeCommon{Lineno: lineno},
		}
		lineno = p.stream.Current().Lineno
	}

	return expr1, nil
}

func (p *parser) parseOr() (nodes.Node, error) {
	lineno := p.stream.Current().Lineno
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.stream.SkipIf("name:or") {
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = makeBinaryOpNode(left, right, "or", lineno)
		lineno = p.stream.Current().Lineno
	}
	return left, nil
}

func (p *parser) parseAnd() (nodes.Node, error) {
	lineno := p.stream.Current().Lineno
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	for p.stream.SkipIf("name:and") {
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = makeBinaryOpNode(left, right, "and", lineno)
		lineno = p.stream.Current().Lineno
	}
	return left, nil
}

func (p *parser) parseNot() (nodes.Node, error) {
	if p.stream.Current().Test("name:not") {
		lineno := p.stream.Next().Lineno
		n, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &nodes.UnaryOp{
			Node:       n,
			Op:         "not",
			NodeCommon: nodes.NodeCommon{Lineno: lineno},
		}, nil
	}
	return p.parseCompare()
}

func (p *parser) parseCompare() (nodes.Node, error) {
	lineno := p.stream.Current().Lineno
	expr, err := p.parseMath1()
	if err != nil {
		return nil, err
	}
	var ops []nodes.Operand

	addOperand := func(tokenType string) error {
		e, err := p.parseMath1()
		if err != nil {
			return err
		}
		ops = append(ops, nodes.Operand{
			Op:         tokenType,
			Expr:       e,
			NodeCommon: nodes.NodeCommon{Lineno: lineno},
		})
		return nil
	}

	for {
		tokenType := p.stream.Current().Type
		if compareOperators.Has(tokenType) {
			p.stream.Next()
			if err := addOperand(tokenType); err != nil {
				return nil, err
			}
		} else if p.stream.SkipIf("name:in") {
			if err := addOperand("in"); err != nil {
				return nil, err
			}
		} else if p.stream.Current().Test("name:not") && p.stream.Look().Test("name:in") {
			p.stream.Skip(2)
			if err := addOperand("notin"); err != nil {
				return nil, err
			}
		} else {
			break
		}
		lineno = p.stream.Current().Lineno
	}

	if len(ops) == 0 {
		return expr, nil
	}
	return &nodes.Compare{
		Expr:       expr,
		Ops:        ops,
		NodeCommon: nodes.NodeCommon{Lineno: lineno},
	}, nil
}

func (p *parser) parseMath1() (nodes.Node, error) {
	lineno := p.stream.Current().Lineno
	left, err := p.parseConcat()
	if err != nil {
		return nil, err
	}

	for p.stream.Current().Type == lexer.TokenAdd || p.stream.Current().Type == lexer.TokenSub {
		currentType := p.stream.Current().Type
		p.stream.Next()
		right, err := p.parseConcat()
		if err != nil {
			return nil, err
		}
		left = makeBinaryOpNode(left, right, currentType, lineno)
		lineno = p.stream.Current().Lineno
	}
	return left, nil
}

func (p *parser) parseMath2() (nodes.Node, error) {
	// TODO it's almost identical as parseMath1
	lineno := p.stream.Current().Lineno
	left, err := p.parsePow()
	if err != nil {
		return nil, err
	}

	for slices.Contains([]string{lexer.TokenMul, lexer.TokenDiv, lexer.TokenFloordiv, lexer.TokenMod}, p.stream.Current().Type) {
		currentType := p.stream.Current().Type
		p.stream.Next()
		right, err := p.parsePow()
		if err != nil {
			return nil, err
		}
		left = makeBinaryOpNode(left, right, currentType, lineno)
		lineno = p.stream.Current().Lineno
	}
	return left, nil

}

func (p *parser) parseConcat() (nodes.Node, error) {
	lineno := p.stream.Current().Lineno
	left, err := p.parseMath2()
	if err != nil {
		return nil, err
	}
	args := []nodes.Node{left}

	for p.stream.Current().Type == lexer.TokenTilde {
		p.stream.Next()
		right, err := p.parseMath2()
		if err != nil {
			return nil, err
		}
		args = append(args, right)
	}
	if len(args) == 1 {
		return args[0], nil
	}
	return &nodes.Concat{
		Nodes:      args,
		NodeCommon: nodes.NodeCommon{Lineno: lineno},
	}, nil
}

func (p *parser) parsePow() (nodes.Node, error) {
	lineno := p.stream.Current().Lineno
	left, err := p.parseUnary(true)
	if err != nil {
		return nil, err
	}

	for p.stream.Current().Type == lexer.TokenPow {
		p.stream.Next()
		right, err := p.parseUnary(true)
		if err != nil {
			return nil, err
		}
		left = makeBinaryOpNode(left, right, lexer.TokenPow, lineno)
		lineno = p.stream.Current().Lineno
	}
	return left, nil
}

func (p *parser) parseUnary(withFilter bool) (node nodes.Node, err error) {
	lineno := p.stream.Current().Lineno
	tokenType := p.stream.Current().Type

	if tokenType == lexer.TokenSub || tokenType == lexer.TokenAdd {
		p.stream.Next()
		node, err = p.parseUnary(false)
		if err != nil {
			return
		}
		node = &nodes.UnaryOp{
			Node:       node,
			Op:         tokenType,
			NodeCommon: nodes.NodeCommon{Lineno: lineno},
		}
	} else {
		node, err = p.parsePrimary()
		if err != nil {
			return
		}
	}

	node, err = p.parsePostfix(node)
	if err != nil {
		return
	}

	if withFilter {
		node, err = p.parseFilterExpr(node)
	}

	return
}

func (p *parser) parsePostfix(node nodes.Node) (nodes.Node, error) {
	var err error

	for {
		tokenType := p.stream.Current().Type
		if tokenType == lexer.TokenDot || tokenType == lexer.TokenLBracket {
			node, err = p.parseSubscript(node)
			if err != nil {
				return nil, err
			}
		} else if tokenType == lexer.TokenLParen {
			node, err = p.parseCall(node)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return node, nil
}

func (p *parser) parseFilterExpr(node nodes.Node) (nodes.Node, error) {
	var err error

	for {
		tokenType := p.stream.Current().Type
		if tokenType == lexer.TokenPipe {
			node, err = p.parseFilter(node, false)
			if err != nil {
				return nil, err
			}
		} else if tokenType == lexer.TokenName && p.stream.Current().Value == "is" {
			node, err = p.parseTest(node)
			if err != nil {
				return nil, err
			}
		} else if tokenType == lexer.TokenLParen {
			node, err = p.parseCall(node)
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	return node, nil
}

func (p *parser) parseSubscript(node nodes.Node) (nodes.Node, error) {
	token := p.stream.Next()
	var arg nodes.Node

	if token.Type == lexer.TokenDot {
		attrToken := p.stream.Current()
		p.stream.Next()
		if attrToken.Type == lexer.TokenName {
			return &nodes.Getattr{
				Node:       node,
				Attr:       attrToken.Value.(string),
				Ctx:        "load",
				NodeCommon: nodes.NodeCommon{Lineno: attrToken.Lineno},
			}, nil
		} else if attrToken.Type != lexer.TokenInteger {
			return nil, p.fail(fmt.Sprintf("expected name or number, got %s", attrToken.Type), &attrToken.Lineno)
		}
		arg = &nodes.Const{
			Value:      attrToken.Value,
			NodeCommon: nodes.NodeCommon{Lineno: attrToken.Lineno},
		}
		return &nodes.Getitem{
			Node:       node,
			Arg:        arg,
			Ctx:        "load",
			NodeCommon: nodes.NodeCommon{Lineno: attrToken.Lineno},
		}, nil
	} else if token.Type == lexer.TokenLBracket {
		var args []nodes.Node
		for p.stream.Current().Type != lexer.TokenRBracket {
			if len(args) > 0 {
				if _, err := p.stream.Expect(lexer.TokenComma); err != nil {
					return nil, err
				}
			}
			arg, err := p.parseSubscribed()
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
		if _, err := p.stream.Expect(lexer.TokenRBracket); err != nil {
			return nil, err
		}

		if len(args) == 1 {
			arg = args[0]
		} else {
			arg = &nodes.Tuple{
				Items:      args,
				Ctx:        "load",
				NodeCommon: nodes.NodeCommon{Lineno: token.Lineno},
			}
		}

		return &nodes.Getitem{
			Node:       node,
			Arg:        arg,
			Ctx:        "load",
			NodeCommon: nodes.NodeCommon{Lineno: token.Lineno},
		}, nil
	}

	return nil, p.fail("expected subscript expression", &token.Lineno)
}

func (p *parser) parseSubscribed() (nodes.Node, error) {
	lineno := p.stream.Current().Lineno
	var args []*nodes.Node

	if p.stream.Current().Type == lexer.TokenColon {
		p.stream.Next()
		args = []*nodes.Node{nil}
	} else {
		node, err := p.parseExpression(true)
		if err != nil {
			return nil, err
		}
		if p.stream.Current().Type != lexer.TokenColon {
			return node, nil
		}
		p.stream.Next()
		args = []*nodes.Node{&node}
	}

	if p.stream.Current().Type == lexer.TokenColon {
		args = append(args, nil)
	} else if p.stream.Current().Type != lexer.TokenRBracket && p.stream.Current().Type != lexer.TokenComma {
		arg, err := p.parseExpression(true)
		if err != nil {
			return nil, err
		}
		args = append(args, &arg)
	} else {
		args = append(args, nil)
	}

	if p.stream.Current().Type == lexer.TokenColon {
		p.stream.Next()
		if p.stream.Current().Type != lexer.TokenRBracket && p.stream.Current().Type != lexer.TokenComma {
			arg, err := p.parseExpression(true)
			if err != nil {
				return nil, err
			}
			args = append(args, &arg)
		} else {
			args = append(args, nil)
		}
	} else {
		args = append(args, nil)
	}

	var start, stop, step *nodes.Node
	if len(args) > 0 {
		start = args[0]
	}
	if len(args) > 1 {
		stop = args[1]
	}
	if len(args) > 2 {
		step = args[2]
	}
	return &nodes.Slice{
		Start:      start,
		Stop:       stop,
		Step:       step,
		NodeCommon: nodes.NodeCommon{Lineno: lineno},
	}, nil
}

func (p *parser) parseCall(node nodes.Node) (nodes.Node, error) {
	lineno := p.stream.Current().Lineno
	args, kwargs, dynArgs, dynKwargs, err := p.parseCallArgs()
	if err != nil {
		return nil, err
	}
	return &nodes.Call{
		Node:       node,
		Args:       args,
		Kwargs:     kwargs,
		DynArgs:    dynArgs,
		DynKwargs:  dynKwargs,
		NodeCommon: nodes.NodeCommon{Lineno: lineno},
	}, nil
}

func (p *parser) parseCallArgs() (args []nodes.Node, kwargs []nodes.Node, dynArgs *nodes.Node, dynKwargs *nodes.Node, err error) {
	var token *lexer.Token
	token, err = p.stream.Expect(lexer.TokenLParen)
	if err != nil {
		return
	}

	requireComma := false

	ensure := func(expr bool) error {
		if !expr {
			return p.fail("invalid syntax for function call expression", &token.Lineno)
		}
		return nil
	}

	for p.stream.Current().Type != lexer.TokenRParen {
		if requireComma {
			if _, err = p.stream.Expect(lexer.TokenComma); err != nil {
				return
			}
			// support for trailing comma
			if p.stream.Current().Type == lexer.TokenRParen {
				break
			}
		}

		var expr nodes.Node
		if p.stream.Current().Type == lexer.TokenMul {
			if err = ensure(dynArgs == nil && dynKwargs == nil); err != nil {
				return
			}
			p.stream.Next()
			expr, err = p.parseExpression(true)
			if err != nil {
				return
			}
			dynArgs = &expr
		} else if p.stream.Current().Type == lexer.TokenPow {
			if err = ensure(dynKwargs == nil); err != nil {
				return
			}
			p.stream.Next()
			expr, err = p.parseExpression(true)
			if err != nil {
				return
			}
			dynKwargs = &expr
		} else {
			if p.stream.Current().Type == lexer.TokenName && p.stream.Look().Type == lexer.TokenAssign {
				// Parsing a kwarg
				if err = ensure(dynKwargs == nil); err != nil {
					return
				}
				key := p.stream.Current().Value
				p.stream.Skip(2)
				expr, err = p.parseExpression(true)
				if err != nil {
					return
				}
				kwargs = append(kwargs, &nodes.Keyword{
					Key:        key.(string),
					Value:      expr,
					NodeCommon: nodes.NodeCommon{Lineno: expr.GetLineno()},
				})
			} else {
				// Parsing an arg
				if err = ensure(dynArgs == nil && dynKwargs == nil && len(kwargs) == 0); err != nil {
					return
				}
				expr, err = p.parseExpression(true)
				if err != nil {
					return
				}
				args = append(args, expr)
			}
		}

		requireComma = true
	}

	_, err = p.stream.Expect(lexer.TokenRParen)
	return
}

func (p *parser) parseFilter(node nodes.Node, startInline bool) (nodes.Node, error) {
	for p.stream.Current().Type == lexer.TokenPipe || startInline {
		if !startInline {
			p.stream.Next()
		}

		token, err := p.stream.Expect(lexer.TokenName)
		if err != nil {
			return nil, err
		}
		name := token.Value.(string)
		for p.stream.Current().Type == lexer.TokenDot {
			p.stream.Next()
			token, err = p.stream.Expect(lexer.TokenName)
			if err != nil {
				return nil, err
			}
			name += "." + token.Value.(string)
		}

		var args []nodes.Node
		var kwargs []nodes.Node
		var dynArgs *nodes.Node
		var dynKwargs *nodes.Node
		if p.stream.Current().Type == lexer.TokenLParen {
			args, kwargs, dynArgs, dynKwargs, err = p.parseCallArgs()
			if err != nil {
				return nil, err
			}
		}

		node = &nodes.Filter{
			Node:       node,
			Name:       name,
			Args:       args,
			Kwargs:     kwargs,
			DynArgs:    dynArgs,
			DynKwargs:  dynKwargs,
			NodeCommon: nodes.NodeCommon{Lineno: token.Lineno},
		}

		startInline = false
	}
	return node, nil
}

func (p *parser) parseTest(node nodes.Node) (nodes.Node, error) {
	// TODO
	panic("not implemented")
}

func (p *parser) parseList() (nodes.Node, error) {
	// TODO
	panic("not implemented")
}

func (p *parser) parseDict() (nodes.Node, error) {
	// TODO
	panic("not implemented")
}

func (p *parser) isTupleEnd(extraEndRules []string) bool {
	current := p.stream.Current()
	if slices.Contains([]string{lexer.TokenVariableEnd, lexer.TokenBlockEnd, lexer.TokenRParen}, current.Type) {
		return true
	} else if extraEndRules != nil {
		return current.TestAny(extraEndRules...)
	}
	return false
}

func (p *parser) parseStatement() ([]nodes.Node, error) {
	token := p.stream.Current()
	if token.Type != lexer.TokenName {
		return nil, p.fail("tag name expected", &token.Lineno)
	}
	p.tagStack.Push(token.Value.(string))
	popTag := true
	defer func() {
		if popTag {
			p.tagStack.Pop()
		}
	}()

	wrapInSlice := func(f func() (nodes.Node, error)) ([]nodes.Node, error) {
		n, err := f()
		if err != nil {
			return nil, err
		}
		return []nodes.Node{n}, nil
	}

	if tokenValue, ok := token.Value.(string); ok {
		if statementKeywords.Has(tokenValue) {
			switch tokenValue {
			case "for":
				return wrapInSlice(p.parseFor)
			case "if":
				return wrapInSlice(p.parseIf)
			case "block":
				return wrapInSlice(p.parseBlock)
			case "extends":
				return wrapInSlice(p.parseExtends)
			case "print":
				return wrapInSlice(p.parsePrint)
			case "macro":
				return wrapInSlice(p.parseMacro)
			case "include":
				return wrapInSlice(p.parseInclude)
			case "from":
				return wrapInSlice(p.parseFrom)
			case "import":
				return wrapInSlice(p.parseImport)
			case "set":
				return wrapInSlice(p.parseSet)
			case "with":
				return wrapInSlice(p.parseWith)
			case "autoescape":
				return wrapInSlice(p.parseAutoescape)
			default:
				panic("unexpected statement keyword " + tokenValue)
			}
		}
	}

	if token.Value == "call" {
		return p.parseCallBlock()
	} else if token.Value == "filter" {
		b, err := p.parseFilterBlock()
		if err != nil {
			return nil, err
		}
		ret := make([]nodes.Node, 1)
		ret[0] = b
		return ret, nil
	} else if tokenValue, ok := token.Value.(string); ok {
		if ext, ok := p.extensions[tokenValue]; ok {
			return ext(p)
		}
	}

	// did not work out, remove the token we pushed by accident
	// from the stack so that the unknown tag fail function can
	// produce a proper error message.
	p.tagStack.Pop()
	popTag = false
	return nil, p.failUnknownTag(token.Value.(string), &token.Lineno)
}

func (p *parser) parseFor() (nodes.Node, error) {
	// TODO
	panic("not implemented")
}

func (p *parser) parseIf() (nodes.Node, error) {
	tok, err := p.stream.Expect("name:if")
	if err != nil {
		return nil, err
	}
	node := &nodes.If{
		NodeCommon: nodes.NodeCommon{Lineno: tok.Lineno},
	}
	result := node

	for {
		node.Test, err = p.parseTuple(false, false, nil, false)
		if err != nil {
			return nil, err
		}
		node.Body, err = p.parseStatements([]string{"name:elif", "name:else", "name:endif"}, false)
		if err != nil {
			return nil, err
		}
		node.Elif = []nodes.If{}
		node.Else = []nodes.Node{}
		token := p.stream.Next()
		if token.Test("name:elif") {
			node = &nodes.If{
				NodeCommon: nodes.NodeCommon{Lineno: token.Lineno},
			}
			result.Elif = append(result.Elif, *node)
			continue
		} else if token.Test("name:else") {
			result.Else, err = p.parseStatements([]string{"name:endif"}, true)
			if err != nil {
				return nil, err
			}
		}
		break
	}

	return result, nil
}

func (p *parser) parseStatements(endTokens []string, dropNeedle bool) ([]nodes.Node, error) {
	p.stream.SkipIf(lexer.TokenColon)
	if _, err := p.stream.Expect(lexer.TokenBlockEnd); err != nil {
		return nil, err
	}
	result, err := p.subparse(endTokens)
	if err != nil {
		return nil, err
	}

	// we reached the end of the template too early, the subparser
	// does not check for this, so we do that now
	if p.stream.Current().Type == lexer.TokenEOF {
		return nil, p.failEOF(endTokens, nil)
	}

	if dropNeedle {
		p.stream.Next()
	}

	return result, nil
}

func (p *parser) parseBlock() (nodes.Node, error) {
	// TODO
	panic("not implemented")
}

func (p *parser) parseExtends() (nodes.Node, error) {
	node := &nodes.Extends{
		NodeCommon: nodes.NodeCommon{Lineno: p.stream.Next().Lineno},
	}
	var err error
	node.Template, err = p.parseExpression(true)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (p *parser) parsePrint() (nodes.Node, error) {
	node := &nodes.Output{
		NodeCommon: nodes.NodeCommon{Lineno: p.stream.Next().Lineno},
		Nodes:      make([]nodes.Node, 0),
	}
	for p.stream.Current().Type != lexer.TokenBlockEnd {
		if len(node.Nodes) > 0 {
			_, err := p.stream.Expect(lexer.TokenComma)
			if err != nil {
				return nil, err
			}
		}
		n, err := p.parseExpression(true)
		if err != nil {
			return nil, err
		}
		node.Nodes = append(node.Nodes, n)
	}
	return node, nil
}

func (p *parser) parseMacro() (nodes.Node, error) {
	n := &nodes.Macro{NodeCommon: nodes.NodeCommon{Lineno: p.stream.Next().Lineno}}

	name, err := p.parseAssignTarget(true, true, nil, false)
	if err != nil {
		return nil, err
	}
	n.Name = name.Name
	if err = p.parseSignature(&n.MacroCall); err != nil {
		return nil, err
	}
	n.Body, err = p.parseStatements([]string{"name:endmacro"}, true)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (p *parser) parseInclude() (nodes.Node, error) {
	// TODO
	panic("not implemented")
}

func (p *parser) parseFrom() (nodes.Node, error) {
	// TODO
	panic("not implemented")
}

func (p *parser) parseImport() (nodes.Node, error) {
	// TODO
	panic("not implemented")
}

func (p *parser) parseSet() (nodes.Node, error) {
	// TODO
	panic("not implemented")
}

func (p *parser) parseWith() (nodes.Node, error) {
	// TODO
	panic("not implemented")
}

func (p *parser) parseAutoescape() (nodes.Node, error) {
	node := &nodes.ScopedEvalContextModifier{
		EvalContextModifier: nodes.EvalContextModifier{
			Options:    make([]nodes.Keyword, 1),
			NodeCommon: nodes.NodeCommon{Lineno: p.stream.Next().Lineno},
		},
	}
	optsExpr, err := p.parseExpression(true)
	if err != nil {
		return nil, err
	}
	node.Options[0] = nodes.Keyword{
		Key:        "autoescape",
		Value:      optsExpr,
		NodeCommon: nodes.NodeCommon{Lineno: optsExpr.GetLineno()},
	}
	node.Body, err = p.parseStatements([]string{"name:endautoescape"}, true)
	if err != nil {
		return nil, err
	}
	return &nodes.Scope{
		Body: []nodes.Node{node},
	}, nil
}

func (p *parser) parseCallBlock() ([]nodes.Node, error) {
	// TODO
	panic("not implemented")
}

func (p *parser) parseFilterBlock() (nodes.Node, error) {
	node := &nodes.FilterBlock{
		NodeCommon: nodes.NodeCommon{Lineno: p.stream.Next().Lineno},
	}

	var err error
	node.Filter, err = p.parseFilter(nil, true)
	if err != nil {
		return nil, err
	}
	node.Body, err = p.parseStatements([]string{"name:endfilter"}, true)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (p *parser) parseAssignTarget(withTuple bool, nameOnly bool, extraEndRule []string, withNamespace bool) (nodes.Name, error) {
	// TODO
	panic("TODO")
}

func (p *parser) parseSignature(n *nodes.MacroCall) error {
	n.Args = make([]nodes.Name, 0)
	n.Defaults = make([]nodes.Expr, 0)
	if _, err := p.stream.Expect(lexer.TokenLParen); err != nil {
		return err
	}
	for p.stream.Current().Type != lexer.TokenRParen {
		if len(n.Args) != 0 {
			if _, err := p.stream.Expect(lexer.TokenComma); err != nil {
				return err
			}
		}
		arg, err := p.parseAssignTarget(true, true, nil, false)
		if err != nil {
			return err
		}
		arg.Ctx = "param"
		if p.stream.SkipIf(lexer.TokenAssign) {
			expr, err := p.parseExpression(true)
			if err != nil {
				return err
			}
			n.Defaults = append(n.Defaults, expr)
		} else if len(n.Defaults) != 0 {
			return p.fail("non-default argument follows default argument", nil)
		}
		n.Args = append(n.Args, arg)
	}
	_, err := p.stream.Expect(lexer.TokenRParen)
	return err
}

func (p *parser) fail(msg string, lineno *int) error {
	var lineNumber int
	if lineno == nil {
		lineNumber = p.stream.Current().Lineno
	} else {
		lineNumber = *lineno
	}
	return errors.TemplateSyntaxError(msg, lineNumber, p.name, p.filename)
}

func (p *parser) failUnknownTag(name string, lineno *int) error {
	return p.failUtEof(&name, p.endTokenStack, lineno)
}

func (p *parser) failEOF(endTokens []string, lineno *int) error {
	if endTokens != nil {
		p.endTokenStack.Push(endTokens)
	}
	return p.failUtEof(nil, p.endTokenStack, lineno)
}

func (p *parser) failUtEof(name *string, endTokenStack *stack.Stack[[]string], lineno *int) error {
	endTokenStackSlice := endTokenStack.Iter()
	expected := set.New[string]()

	for _, exprs := range endTokenStackSlice {
		for _, expr := range exprs {
			expected.Add(lexer.DescribeTokenExpr(expr))
		}
	}

	var currentlyLooking *string
	if len(endTokenStackSlice) > 0 {
		lastEndToken := endTokenStackSlice[len(endTokenStackSlice)-1]
		var described []string
		for _, expr := range lastEndToken {
			described = append(described, fmt.Sprintf("%q", lexer.DescribeTokenExpr(expr)))
		}
		v := strings.Join(described, " or ")
		currentlyLooking = &v
	}

	var messages []string
	if name == nil {
		messages = append(messages, "Unexpected end of template.")
	} else {
		messages = append(messages, fmt.Sprintf("Encountered unknown tag %q.", *name))
	}

	if currentlyLooking != nil {
		if name != nil && expected.Has(*name) {
			messages = append(messages, "You probably made a nesting mistake. Jinja is expecting this tag, but currently looking for "+*currentlyLooking+".")
		} else {
			messages = append(messages, "Jinja was looking for the following tags: "+*currentlyLooking+".")
		}
	}

	if lastTag := p.tagStack.Peek(); lastTag != nil {
		messages = append(messages, fmt.Sprintf("The innermost block that needs to be closed is %q.", *lastTag))
	}

	return p.fail(strings.Join(messages, " "), lineno)
}
