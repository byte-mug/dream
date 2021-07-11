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


package parsex

import "github.com/byte-mug/semiparse/parser"
import "github.com/byte-mug/semiparse/scanlist"

type Snip struct {
	Inner parser.ParseRule
}
func (c Snip) Parse(p *parser.Parser, tokens *scanlist.Element, left interface{}) parser.ParserResult {
	res := c.Inner.Parse(p,tokens,left)
	if res.Result == parser.RESULT_FAILED { res.Result = parser.RESULT_FAILED_CUT }
	return res
}

func Jump() parser.ParserResult { return parser.ParserResult{ Result: parser.RESULT_FAILED, Data:"You should not be seeing this!" } }

func DoCut(res parser.ParserResult) parser.ParserResult {
	if res.Result == parser.RESULT_FAILED { res.Result = parser.RESULT_FAILED_CUT }
	return res
}
/*
func DontCut(res parser.ParserResult) parser.ParserResult {
	if res.Result == parser.RESULT_FAILED_CUT { res.Result = parser.RESULT_FAILED }
	return res
}
*/

type ArrayBreakIf struct {
	Elem parser.ParseRule
	Term parser.ParseRule
}
func (s ArrayBreakIf) Parse(p *parser.Parser, tokens *scanlist.Element, left interface{}) parser.ParserResult {
	dok := []interface{}{}
	for {
		if s.Term.Parse(p,tokens,nil).Ok() { return parser.ResultOk(tokens,dok) }
		npr := s.Elem.Parse(p,tokens,nil)
		if !npr.Ok() { return npr }
		dok = append(dok,npr.Data)
		tokens = npr.Next
	}
	panic("unreachable")
}

type ArrayDelimited struct {
	Elem parser.ParseRule
	Delim parser.ParseRule
}
func (s ArrayDelimited) Parse(p *parser.Parser, tokens *scanlist.Element, left interface{}) parser.ParserResult {
	dok := []interface{}{nil}
	{
		npr := s.Elem.Parse(p,tokens,nil)
		if !npr.Ok() { return npr }
		dok[0] = npr.Data
		tokens = npr.Next
	}
	for {
		npr := s.Delim.Parse(p,tokens,nil)
		if !npr.Ok() { return parser.ResultOk(tokens,dok) }
		npr = s.Elem.Parse(p,npr.Next,nil)
		if !npr.Ok() { return npr }
		dok = append(dok,npr.Data)
		tokens = npr.Next
	}
	panic("unreachable")
}


type DelegateShort string
func (d DelegateShort) Parse(p *parser.Parser,tokens *scanlist.Element, left interface{}) (parser.ParserResult) {
	return p.MatchNoLeftRecursion(string(d),tokens)
}
