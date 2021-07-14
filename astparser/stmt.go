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

import "text/scanner"
import "github.com/byte-mug/dream/parsex"
import "github.com/byte-mug/semiparse/scanlist"
import "github.com/byte-mug/semiparse/parser"
import "fmt"

var declSigil = parser.OR{require('$'),require('@'),require('%')}
var declVarSingle = parser.ArraySeq{declSigil, parser.Pfunc(d_ident)}

var declVarList = parser.ArraySeq{
	declVarSingle,
	parser.ArrayStar{parser.LSeq{require(','),declVarSingle}},
}
var stmt_block_o = require('{')
var stmt_block_c = require('}')


func d_declvarlist(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := declVarList.Parse(p,tokens,left)
	if !res.Ok() { return res }
	
	pair := res.Data.([]interface{})
	pair = append(pair[:1],pair[1].([]interface{})...)
	for i := range pair {
		pair[i] = fmt.Sprint(pair[i].([]interface{})...)
	}
	res.Data = pair
	
	return res
}

var declMy = parser.ArraySeq{require(KW_my), parser.Pfunc(d_declvarlist), require(';')}

func d_decl_myvar(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if tokens==nil { return parser.ResultFail("EOF!",scanner.Position{}) }
	res := declMy.Parse(p,tokens,left)
	if !res.Ok() { return res }
	
	res.Data = &SMyVars{res.Data.([]interface{})[1].([]interface{}),tokens.Pos}
	
	return res
}

func d_stmtsub_expr(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	if stmt_block_o.Parse(p,tokens,left).Ok() { return parsex.Jump() }
	res := p.Match("Expr",tokens)
	if !res.Ok() { return res }
	if IsArrayExpr(res.Data) {
		res.Data = &SArray{res.Data,tokens.Pos}
	} else {
		res.Data = &SExpr{res.Data,tokens.Pos}
	}
	
	return res
}

var stmtsub_print = parser.LSeq{
	parser.RequireText{"print"},
	parsex.Snip{parser.Delegate("Expr")},
}
func d_stmtsub_print(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := stmtsub_print.Parse(p,tokens,left)
	if res.Ok() { res.Data = &SPrint{res.Data,tokens.Pos} }
	return res
}


var stmtsub_cond = parser.ArraySeq{
	parser.OR{require(KW_if),require(KW_unless),require(KW_while),require(KW_for)},
	parsex.Snip{parser.Delegate("Expr")},
}
func d_stmtsub_cond(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := stmtsub_cond.Parse(p,tokens,left)
	if !res.Ok() { return res }
	r := res.Data.([]interface{})
	op := r[0].(string)
	if op=="for" {
		res.Data = &SFor{"_",r[1],left,tokens.Pos}
	} else {
		res.Data = &SCond{op,r[1],left,tokens.Pos}
	}
	
	return res
}


var stmt_sub = parser.ArraySeq{parser.Delegate("StmtSub"),parsex.Snip{require(';')}}
func d_stmt_sub(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := stmt_sub.Parse(p,tokens,left)
	if res.Ok() { res.Data = res.Data.([]interface{})[0] }
	return res
}

var stmt_block = parser.ArraySeq{
	stmt_block_o,
	parsex.ArrayBreakIf{parsex.Snip{parser.Delegate("Stmt")},stmt_block_c},
	stmt_block_c,
}
func d_stmt_block(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := stmt_block.Parse(p,tokens,nil)
	if !res.Ok() { return res }
	res.Data = &SBlock{res.Data.([]interface{})[1].([]interface{}),tokens.Pos}
	return res
}

var stmt_cond = parser.ArraySeq{
	parser.OR{require(KW_if),require(KW_unless),require(KW_while)},
	parsex.Snip{require('(')},
	parsex.Snip{parser.Delegate("Expr")},
	parsex.Snip{require(')')},
	parsex.Snip{parser.Delegate("Stmt")},
}
var possibleElse = map[string]bool { "if":true, "unless":true, }
var stmt_cond_haselse = require(KW_else)
var stmt_cond_suffix = parser.LSeq{
	stmt_cond_haselse,
	parsex.Snip{parser.Delegate("Stmt")},
}

func d_stmt_cond(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := stmt_cond.Parse(p,tokens,left)
	if !res.Ok() { return res }
	r := res.Data.([]interface{})
	op := r[0].(string)
	tokens = res.Next
	if possibleElse[op] && stmt_cond_haselse.Parse(p,tokens,nil).Ok() {
		res2 := stmt_cond_suffix.Parse(p,tokens,nil)
		if !res2.Ok() { return res2 }
		res.Data = &SIfElse{op, r[2], r[4], res2.Data, tokens.Pos}
		res.Next = res2.Next
	} else {
		res.Data = &SCond{op, r[2], r[4], tokens.Pos}
	}
	return res
}

var stmt_for1 = parser.ArraySeq{
	require(KW_for), // -2
	require('$'),//-1
	parsex.Snip{parser.Pfunc(d_ident)}, // 0
	parsex.Snip{require('(')},
	parsex.Snip{parser.Delegate("Expr")}, // 2
	parsex.Snip{require(')')},
	parsex.Snip{parser.Delegate("Stmt")}, // 4
}

var stmt_for2 = parser.ArraySeq{
	require(KW_for), // 0 (overwritten)
	parsex.Snip{require('(')},
	parsex.Snip{parser.Delegate("Expr")}, // 2
	parsex.Snip{require(')')},
	parsex.Snip{parser.Delegate("Stmt")}, // 4
}

var stmt_for = parser.OR{stmt_for1,stmt_for2}

func d_stmt_for(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := stmt_for.Parse(p,tokens,left)
	if !res.Ok() { return res }
	r := res.Data.([]interface{})
	
	if len(r)==7 {
		r = r[2:]
	} else {
		r[0] = "_"
	}
	
	res.Data = &SFor{r[0].(string), r[2], r[4], tokens.Pos}
	
	return res
}

var stmt_semicolon = require(';')
func d_stmt_semicolon(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := stmt_semicolon.Parse(p,tokens,left)
	if res.Ok() { res.Data = &SNoop{tokens.Pos} }
	return res
}


func RegisterStmt(p *parser.Parser) {
	p.Define("Decl",false,parser.Pfunc(d_decl_myvar))
	
	p.Define("StmtSub",false,parser.Pfunc(d_stmtsub_expr))
	p.Define("StmtSub",false,parser.Pfunc(d_stmtsub_print))
	p.Define("StmtSub",true,parser.Pfunc(d_stmtsub_cond))
	
	p.Define("Stmt",false,parser.Pfunc(d_stmt_semicolon))
	p.Define("Stmt",false,parser.Pfunc(d_stmt_cond))
	p.Define("Stmt",false,parser.Pfunc(d_stmt_for))
	p.Define("Stmt",false,parser.Pfunc(d_stmt_block))
	p.Define("Stmt",false,parser.Pfunc(d_stmt_sub))
	
	p.Define("Stmt",false,parser.Delegate("Decl"))
}

