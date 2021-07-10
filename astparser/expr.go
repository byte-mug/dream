/*
Copyright (c) 2021 Simon Schmidt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/


package astparser

import "strconv"
import "github.com/byte-mug/dream/values"
import "text/scanner"
import "github.com/byte-mug/semiparse/scanlist"
import "github.com/byte-mug/semiparse/parser"
import "regexp"
import "fmt"

func textify(r rune) string {
	if KW_min_>r && r>KW_max_ {
		for v,k := range Keywords {
			if k==r { return v }
		}
		return "<?>"
	}
	return parser.Textify(r)
}

func require(r rune) parser.ParseRule {
	return parser.Required{ r, textify }
}

func d_ident(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens==nil { return parser.ResultFail("EOF!",scanner.Position{}) }
	switch {
	case tokens.Token == scanner.Ident, KW_min_>tokens.Token && tokens.Token>KW_max_:
		return parser.ResultOk(tokens.Next(), tokens.TokenText )
	}
	return parser.ResultFail("Invalid Expression!",tokens.Pos)
}

var vssigil = require('$')
var vsprefix = parser.OR{
	parser.Pfunc(d_ident),
	require(scanner.Int),
	parser.Delegate("Vscalar"),
	parser.ArraySeq{require('{'), parser.Delegate("Expr"), require('}')},
}
var vssuffix = parser.OR{
	parser.ArraySeq{require('{'), parser.Pfunc(d_ident), require('}')},
	parser.ArraySeq{require('{'), parser.Delegate("Expr"), require('}')},
	parser.ArraySeq{require('['), parser.Delegate("Expr"), require(']')},
}

func d_vscalar(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens==nil { return parser.ResultFail("EOF!",scanner.Position{}) }
	
	pos := tokens.Pos
	
	// $
	res1 := vssigil.Parse(p,tokens,left)
	if !res1.Ok() { return res1 }
	tokens = res1.Next
	
	// $THIS_PART
	res1 = vsprefix.Parse(p,tokens,left)
	if !res1.Ok() { return res1 }
	tokens = res1.Next
	
	// Process data
	var src interface{}
	switch v := res1.Data.(type) {
	case string: src = v // d_ident | int
	case []interface{}: src = v[1] // long expr
	default: src = v // short expr
	}
	tokens = res1.Next
	
	// $var{SUFFIX} $var[SUFFIX]
	res1 = vssuffix.Parse(p,tokens,left)
	if !res1.Ok() { return parser.ResultOk(tokens, &EScalar{src, pos}) }
	rl := res1.Data.([]interface{})
	switch rl[0].(string) {
	case "{":/*}*/ return parser.ResultOk(res1.Next, &EHashScalar{src, rl[1] ,pos})
	case "[":/*]*/ return parser.ResultOk(res1.Next, &EArrayScalar{src, rl[1] ,pos})
	}
	
	return parser.ResultFail("Invalid Scalar Expression!",pos)
}

func d_literal(token *scanlist.Element) interface{} {
	var lit values.Scalar = nil
	switch token.Token {
	case KW_undef: lit = values.Null()
	case scanner.Int: i,_ := strconv.ParseInt(token.TokenText, 0, 64); lit = values.ScInt(i)
	case scanner.Float: f,_ := strconv.ParseFloat(token.TokenText, 64); lit = values.ScFloat(f)
	case scanner.Char,scanner.String,scanner.RawString: s,_ := strconv.Unquote(token.TokenText); lit = values.ScString(s)
	}
	if lit==nil { return nil }
	return &ELiteral{lit, token.Pos}
}

func d_expr0(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens==nil { return parser.ResultFail("EOF!",scanner.Position{}) }
	if obj := d_literal(tokens); obj!=nil { return parser.ResultOk(tokens.Next(),obj) }
	
	
	switch tokens.Token {
	//case scanner.Ident: return parser.ResultOk(tokens.Next(),&Expr{E_VAR,tokens.TokenText,nil,tokens.Pos})
	case '+','-','!','~':{
		sub := p.MatchNoLeftRecursion("Expr0",tokens.Next())
		if sub.Result==parser.RESULT_OK {
			sub.Data = &EUnop{tokens.TokenText,sub.Data,tokens.Pos}
		}
		return sub
	    }
	case KW_scalar:{
		sub := p.MatchNoLeftRecursion("Array",tokens.Next())
		if sub.Ok() {
			sub.Data = &EFromArray{sub.Data,tokens.Pos}
		}
	    }
	case '(': /*)*/{
		sub := p.Match("Expr",tokens.Next())
		if sub.Result==parser.RESULT_OK {
			e,t := parser.Match(/*(*/ parser.Textify,sub.Next,')')
			if e!=nil { return parser.ResultFail(fmt.Sprint(e),sub.Next.SafePos()) }
			sub.Next = t
		}
		return sub
	    }
	}
	return parser.ResultFail("Invalid Expression!",tokens.Pos)
}


var vbinop_simple = parser.OR{
	require('+'),
	require('-'),
	require('*'),
	require('/'),
	require('%'),
	require('.'),
}

var vbinop_single = parser.OR{
	vbinop_simple,
	require(KW_and),
	require(KW_or),
	require(KW_eq),
	require(KW_ne),
	require(KW_ge),
	require(KW_gt),
	require(KW_le),
	require(KW_lt),
	require('<'),
	require('>'),
}
var vbinop = parser.OR{
	parser.ArraySeq{require('<'),require('=')},
	parser.ArraySeq{require('>'),require('=')},
	parser.ArraySeq{require('='),require('=')},
	parser.ArraySeq{require('!'),require('=')},
	parser.ArraySeq{vbinop_single},
}

func d_expr0_trailer(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens==nil { return parser.ResultFail("EOF!",scanner.Position{}) }
	
	pos := tokens.Pos
	
	res1 := vbinop.Parse(p,tokens,nil)
	if !res1.Ok() { return res1 }
	tokens = res1.Next
	
	op := fmt.Sprint(res1.Data.([]interface{})...)
	
	res1 = p.MatchNoLeftRecursion("Expr0",tokens)
	
	if res1.Ok() {
		res1.Data = &EBinop{op,left,res1.Data,pos}
	}
	return res1
}

var rxlit = parser.OR{
	require(scanner.Char),
	require(scanner.String),
	require(scanner.RawString),
}

var rxtrail = parser.OR{
	parser.ArraySeq{require('='),require('~'),require(KW_m),rxlit},
	parser.ArraySeq{require('='),require('~'),require(KW_s),rxlit,parser.Delegate("Expr0")},
}

func d_expr1_trailer(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := rxtrail.Parse(p,tokens,nil)
	if !res.Ok() { return res }
	rxx := res.Data.([]interface{})
	rest := res.Next
	
	rxs := rxx[3].(string)
	rxs = rxs[1:len(rxs)-1]
	rx,err := regexp.Compile(rxs)
	if err!=nil { return parser.ResultFail("Invalid regex: "+err.Error(),tokens.Pos) }
	
	
	switch rxx[2].(string) {
	case "m":
		if rest.SafeTokenText()=="g" {
			res.Next = rest.Next()
			res.Data = &EMatchGlobal{left,rx,tokens.Pos}
		} else {
			res.Data = &EMatch{left,rx,tokens.Pos}
		}
	case "s":
		res.Data = &EReplace{left,rx,rxx[4],tokens.Pos}
	}
	return res
}

var assign = parser.ArraySeq{require('='),parser.Delegate("Expr3")}

func d_expr3_trailer1(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := assign.Parse(p,tokens,nil)
	if !res.Ok() { return res }
	res.Data = &EScAssign{left,res.Data.([]interface{})[1],tokens.Pos}
	return res
}


var opassign = parser.ArraySeq{vbinop_simple,require('='),parser.Delegate("Expr3")}

func d_expr3_trailer2(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := opassign.Parse(p,tokens,nil)
	if !res.Ok() { return res }
	res.Data = &EBinopAssign{res.Data.([]interface{})[0].(string),left,res.Data.([]interface{})[2],tokens.Pos}
	return res
}

var vasigil = parser.OR{ require('@'), require('%') }
var vaname = parser.OR{
	parser.Pfunc(d_ident),
	require(scanner.Int),
	parser.Delegate("Vscalar"),
	parser.ArraySeq{require('{'), parser.Delegate("Expr"), require('}')},
}
var vacomplete = parser.ArraySeq{ vasigil,vaname }

func d_array_variable(p *parser.Parser, tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := vacomplete.Parse(p,tokens,nil)
	if !res.Ok() { return res }
	
	list := res.Data.([]interface{})
	switch list[0].(string) {
	case "@": res.Data = &AArray{list[1],tokens.Pos}
	case "%": res.Data = &AHash{list[1],tokens.Pos}
	}
	
	return res
}

var va_assign = parser.LSeq{require('='),parser.Delegate("Array0")}
func d_array_assign(p *parser.Parser, tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := va_assign.Parse(p,tokens,nil)
	if res.Ok() { res.Data = &AArAssign{left,res.Data,tokens.Pos} }
	return res
}

func RegisterExpr(p *parser.Parser) {

	p.Define("Vscalar",false,parser.Pfunc(d_vscalar))
	p.Define("Expr0",false,parser.Delegate("Vscalar"))
	p.Define("Expr0",false,parser.Pfunc(d_expr0))
	p.Define("Expr0",true,parser.Pfunc(d_expr0_trailer))
	
	p.Define("Expr1",false,parser.Delegate("Expr0"))
	p.Define("Expr1",true,parser.Pfunc(d_expr1_trailer))
	
	p.Define("Expr2",false,parser.Delegate("Expr1"))
	
	
	p.Define("Expr3",false,parser.Delegate("Expr2"))
	p.Define("Expr3",true,parser.Pfunc(d_expr3_trailer1))
	p.Define("Expr3",true,parser.Pfunc(d_expr3_trailer2))
	
	p.Define("Expr",false,parser.Delegate("Expr3"))
	
	// ----------------------- Array ----------------------
	
	p.Define("Array0",false,parser.Pfunc(d_array_variable))
	p.Define("Array0",true,parser.Pfunc(d_array_assign))
	
	p.Define("Array",false,parser.Delegate("Array0"))
}

