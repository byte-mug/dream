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

import "github.com/byte-mug/semiparse/scanlist"
import "github.com/byte-mug/semiparse/parser"
import "github.com/byte-mug/dream/parsex"


var mdecl_sub = parser.ArraySeq{
	require(KW_sub),
	parsex.Snip{parser.Pfunc(d_ident)},
	parsex.Snip{parser.Pfunc(d_stmt_block)},
}

func d_mdecl_sub(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := mdecl_sub.Parse(p,tokens,nil)
	if !res.Ok() { return res }
	
	arr := res.Data.([]interface{})
	
	res.Data = &MDSub{arr[1].(string),arr[2],tokens.Pos}
	
	return res
}

type mDecl struct{ I interface{} }

func d_melem_mdecl(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := p.Match("Mdecl",tokens)
	if res.Ok() { res.Data = &mDecl{res.Data} }
	return res
}

var end_of_file = parser.OR{
	parser.RequireText{"__END__"},
}
func d_module(p *parser.Parser,tokens *scanlist.Element, left interface{}) parser.ParserResult {
	m := new(Module)
	var stmts []interface{}
	pos := tokens.Pos
	
	for tokens!=nil {
		if end_of_file.Parse(p,tokens,nil).Ok() { break }
		res := p.Match("Melem",tokens)
		if !res.Ok() { return res }
		tokens = res.Next
		switch e := res.Data.(type) {
		case *mDecl:
			switch t := e.I.(type) {
			case *MDSub: m.Subs = append(m.Subs,t)
			}
		default: stmts = append(stmts,e)
		}
	}
	m.Main = &MDSub{"",&SBlock{stmts,pos},pos}
	return parser.ResultOk(tokens,m)
}

func RegisterModule(p *parser.Parser) {
	p.Define("Mdecl",false,parser.Pfunc(d_mdecl_sub))
	
	p.Define("Melem",false,parser.Pfunc(d_melem_mdecl))
	p.Define("Melem",false,parser.Delegate("Stmt"))
	p.Define("Module",false,parser.Pfunc(d_module))
}

func Register(p *parser.Parser) {
	RegisterExpr(p)
	RegisterStmt(p)
	RegisterModule(p)
}

