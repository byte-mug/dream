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
import "github.com/byte-mug/dream/parsex"
import "github.com/byte-mug/dream/values"
import "text/scanner"
import "github.com/byte-mug/semiparse/scanlist"
import "github.com/byte-mug/semiparse/parser"
import "regexp"
import "strings"
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

var module_name_sep = parser.LSeq{require(':'),require(':')}
var module_name = parsex.ArrayDelimited{parser.Pfunc(d_ident),module_name_sep}
func d_module_name(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := module_name.Parse(p,tokens,left)
	if !res.Ok() { return res }
	
	arr := res.Data.([]interface{})
	var sobj strings.Builder
	for i,elem := range arr {
		if i>0 { sobj.WriteString("::") }
		sobj.WriteString(elem.(string))
	}
	res.Data = sobj.String()
	
	return res
}
var module_name_guess = parser.LSeq{parser.Pfunc(d_ident),require(':'),require(':')}


var vssigil = require('$')
var vsprefix = parser.OR{
	parser.Pfunc(d_ident),
	require(scanner.Int),
	parser.Delegate("Vscalar"),
	parser.ArraySeq{require('{'), parsex.Snip{parser.Delegate("Expr")}, parsex.Snip{require('}')}},
}
var vssuffix = parser.OR{
	parser.ArraySeq{require('{'), parser.Pfunc(d_ident), require('}')},
	parser.ArraySeq{require('{'), parsex.Snip{parser.Delegate("Expr")}, parsex.Snip{require('}')}},
	parser.ArraySeq{require('['), parsex.Snip{parser.Delegate("Expr")}, parsex.Snip{require(']')}},
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
var vscalarspec = parser.LSeq{require('$'),require('@')}
func d_vscalarspec(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := vscalarspec.Parse(p,tokens,nil)
	if res.Ok() { res.Data = &EScalar{res.Data, tokens.Pos} }
	return res
}


func d_literal(token *scanlist.Element) interface{} {
	var lit values.Scalar = nil
	if ok,_ := parser.FastMatch(token,KW_undef,':',':'); ok { return nil }
	switch token.Token {
	case KW_undef: lit = values.Null()
	case scanner.Int: i,_ := strconv.ParseInt(token.TokenText, 0, 64); lit = values.ScInt(i)
	case scanner.Float: f,_ := strconv.ParseFloat(token.TokenText, 64); lit = values.ScFloat(f)
	case scanner.Char,scanner.String,scanner.RawString: s,_ := strconv.Unquote(token.TokenText); lit = values.ScString(s)
	}
	if lit==nil { return nil }
	return &ELiteral{lit, token.Pos}
}

func d_expr0_module_name(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := d_module_name(p,tokens,left)
	if res.Ok() { res.Data = &EModule{res.Data.(string),tokens.Pos} }
	return res
}

func d_expr0(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens==nil { return parser.ResultFail("EOF!",scanner.Position{}) }
	if module_name_guess.Parse(p,tokens,nil).Ok() { return d_expr0_module_name(p,tokens,left) }
	
	if obj := d_literal(tokens); obj!=nil { return parser.ResultOk(tokens.Next(),obj) }
	
	switch tokens.Token {
	case '+','-','!','~':{
		sub := p.MatchNoLeftRecursion("Expr0",tokens.Next())
		if sub.Result==parser.RESULT_OK {
			sub.Data = &EUnop{tokens.TokenText,sub.Data,tokens.Pos}
		}
		return sub
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
	return parser.ResultFail("unexpected "+textify(tokens.Token)+", expected [+-!~]<expr> or (<expr>)!",tokens.Pos)
}

var vsexlist_kv = parser.ArraySeq{
	parser.OR{
		parser.Pfunc(d_ident),
		parsex.DelegateShort("Expr3"),
	},
	require('='),
	require('>'),
	parsex.DelegateShort("Expr3"),
}

func d_exlist_kv(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := vsexlist_kv.Parse(p,tokens,nil)
	if !res.Ok() { return res }
	
	arr := res.Data.([]interface{})
	
	res.Data = &AConcat{[]interface{}{arr[0],arr[3]},tokens.Pos}
	
	return res
}

var vsexlist_elem = parser.OR{
	parser.Pfunc(d_exlist_kv),
	parsex.DelegateShort("Expr3"),
}

var vsexlist = parsex.ArrayDelimited{
	vsexlist_elem,
	require(','),
}

var vsreflit = parser.OR{
	parser.ArraySeq{require('{'),require('}')},
	parser.ArraySeq{require('['),require(']')},
	parser.ArraySeq{require('('),require(')')},
	parser.ArraySeq{require('{'),parsex.Snip{vsexlist},parsex.Snip{require('}')}},
	parser.ArraySeq{require('['),parsex.Snip{vsexlist},parsex.Snip{require(']')}},
	parser.ArraySeq{require('('),parsex.Snip{vsexlist},parsex.Snip{require(')')}},
}
func d_expr0_ref(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := vsreflit.Parse(p,tokens,left)
	if !res.Ok() { return res }
	
	arr := res.Data.([]interface{})
	
	// Avoid cascading *AConcat objects!
	arr = flatten_one_level(arr)
	
	var elems []interface{}
	if len(arr)==3 { elems = arr[1].([]interface{}) }
	switch arr[0].(string) {
	case "{":/*}*/ res.Data = &ECreateHash{elems,tokens.Pos}
	case "[":/*]*/ res.Data = &ECreateArray{elems,tokens.Pos}
	case "(":/*)*/ res.Data = &AConcat{elems,tokens.Pos}
	}
	
	return res
}

var va_subcall = parser.OR{
	parser.ArraySeq{ require(scanner.Ident),require('('),require(')') },
	parser.ArraySeq{ require(scanner.Ident),require('('),parsex.Snip{vsexlist},parsex.Snip{require(')')} },
}

func d_expr0_call(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := va_subcall.Parse(p,tokens,left)
	if !res.Ok() { return res }
	arr := res.Data.([]interface{})
	
	var args []interface{} = nil
	
	if len(arr)==4 {
		args = arr[2].([]interface{})
		// Avoid cascading *AConcat objects!
		args = flatten_one_level(args)
	}
	
	res.Data = &ESubCall{arr[0].(string),args,tokens.Pos}
	
	return res
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
	
	if ok,_ := parser.FastMatch(tokens,'-','>'); ok { return parsex.Jump() }
	
	pos := tokens.Pos
	
	res1 := vbinop.Parse(p,tokens,nil)
	if !res1.Ok() { return res1 }
	tokens = res1.Next
	
	op := fmt.Sprint(res1.Data.([]interface{})...)
	
	res1 = parsex.DoCut(p.MatchNoLeftRecursion("Expr0",tokens))
	
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
	parser.ArraySeq{require('='),require('~'),require(KW_s),rxlit,parsex.DelegateShort("Expr1")},
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

var oparrow = parser.OR{
	parser.ArraySeq{parser.Pfunc(d_ident),parsex.Snip{require('(')},require(')')},
	parser.ArraySeq{parser.Pfunc(d_ident),parsex.Snip{require('(')},parsex.Snip{vsexlist},parsex.Snip{require(')')}},
	parser.ArraySeq{require('{'), parser.Pfunc(d_ident), require('}')},
	parser.ArraySeq{require('{'), parsex.Snip{parser.Delegate("Expr")}, parsex.Snip{require('}')}},
	parser.ArraySeq{require('['), parsex.Snip{parser.Delegate("Expr")}, parsex.Snip{require(']')}},
}

func d_expr1_arrow(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens==nil { return parser.ResultFail("EOF!",scanner.Position{}) }
	pos := tokens.Pos
	var ok bool
	
	ok,tokens = parser.FastMatch(tokens,'-','>')
	if !ok { return parser.ResultFail("not matched",pos) }
	
	res := oparrow.Parse(p,tokens,nil)
	if !res.Ok() { return res }
	arr := res.Data.([]interface{})
	str := arr[0].(string)
	
	switch str {
	case "{":/*}*/ res.Data = &EHashScalar{left, arr[1] ,pos}
	case "[":/*]*/ res.Data = &EArrayScalar{left, arr[1] ,pos}
	default:
		call := &EObjCall{left,str,nil,pos}
		if len(arr)==4 { call.Args = arr[2].([]interface{}) }
		res.Data = call
	}
	
	return res
}

var suffixgo = parser.RequireText{"go"}

func d_expr2_go(p *parser.Parser, tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := suffixgo.Parse(p,tokens,nil)
	if !res.Ok() { return res }
	if _,ok := left.(callExpr); !ok { return parser.ResultFail("unexpected go",tokens.Pos) }
	
	res.Data = &EGoFunction{left,tokens.Pos}
	return res
}

var exprifelse = parser.ArraySeq{
	require('?'),
	parsex.Snip{parser.Delegate("Expr")},
	parsex.Snip{require(':')},
	parsex.Snip{parser.Delegate("Expr")},
}

func d_expr2_ifelse(p *parser.Parser, tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := exprifelse.Parse(p,tokens,nil)
	if !res.Ok() { return res }
	list := res.Data.([]interface{})
	
	if IsArrayExpr(list[1]) || IsArrayExpr(list[3]) {
		res.Data = &AExIfElse{left,list[1],list[3],tokens.Pos}
	} else {
		res.Data = &EExIfElse{left,list[1],list[3],tokens.Pos}
	}
	
	return res
}

var assign = parser.ArraySeq{require('='),parser.Delegate("Expr3")}

func d_expr3_trailer1(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := assign.Parse(p,tokens,nil)
	if !res.Ok() { return res }
	
	if IsArrayExpr(left) {
		res.Data = &AArAssign{left,res.Data.([]interface{})[1],tokens.Pos}
	} else {
		res.Data = &EScAssign{left,res.Data.([]interface{})[1],tokens.Pos}
	}
	
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
	parser.ArraySeq{require('{'), parsex.Snip{parser.Delegate("Expr")}, parsex.Snip{require('}')}},
}
var vacomplete = parser.ArraySeq{ vasigil,vaname }

func d_array_variable(p *parser.Parser, tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := vacomplete.Parse(p,tokens,nil)
	if !res.Ok() { return res }
	
	list := res.Data.([]interface{})
	if arr,ok := list[1].([]interface{}); ok { list[1] = arr[1] }
	switch list[0].(string) {
	case "@": res.Data = &AArray{list[1],tokens.Pos}
	case "%": res.Data = &AHash{list[1],tokens.Pos}
	}
	
	return res
}

func RegisterExpr(p *parser.Parser) {

	p.Define("Vscalar",false,parser.Pfunc(d_vscalar))
	p.Define("Vscalar",false,parser.Pfunc(d_vscalarspec))
	p.Define("Expr0",false,parser.Delegate("Vscalar"))
	p.Define("Expr0",false,parser.Pfunc(d_array_variable))
	p.Define("Expr0",false,parser.Pfunc(d_expr0))
	
	p.Define("Expr0",false,parser.Pfunc(d_expr0_ref))
	p.Define("Expr0",false,parser.Pfunc(d_expr0_call))
	p.Define("Expr0",false,parser.Pfunc(d_expr0_module_name))
	p.Define("Expr0",true,parser.Pfunc(d_expr0_trailer))
	
	p.Define("Expr1",false,parser.Delegate("Expr0"))
	p.Define("Expr1",true,parser.Pfunc(d_expr1_trailer))
	p.Define("Expr1",true,parser.Pfunc(d_expr1_arrow))
	
	p.Define("Expr2",false,parser.Delegate("Expr1"))
	p.Define("Expr2",true,parser.Pfunc(d_expr2_go))
	p.Define("Expr2",true,parser.Pfunc(d_expr2_ifelse))
	
	
	p.Define("Expr3",false,parser.Delegate("Expr2"))
	p.Define("Expr3",true,parser.Pfunc(d_expr3_trailer1))
	p.Define("Expr3",true,parser.Pfunc(d_expr3_trailer2))
	
	p.Define("Expr",false,parser.Delegate("Expr3"))
	
}

