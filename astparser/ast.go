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

import "github.com/byte-mug/dream/values"
import "github.com/byte-mug/semiparse/scanlist"
import "text/scanner"
import "regexp"
import "fmt"

const (
	KW_min_ rune = -(100+iota)
	KW_undef
	KW_and
	KW_or
	KW_eq
	KW_ne
	KW_lt
	KW_le
	KW_gt
	KW_ge
	KW_m
	KW_s
	KW_my
	KW_if
	KW_unless
	KW_while
	KW_else
	KW_scalar
	KW_max_
)

var Keywords = scanlist.TokenDict{
	"undef" : KW_undef,
	"and"   : KW_and,
	"or"    : KW_or,
	"eq"    : KW_eq,
	"ne"    : KW_ne,
	"lt"    : KW_lt,
	"le"    : KW_le,
	"gt"    : KW_gt,
	"ge"    : KW_ge,
	"m"     : KW_m,
	"s"     : KW_s,
	"my"    : KW_my,
	"if"    : KW_if,
	"unless": KW_unless,
	"while" : KW_while,
	"else"  : KW_else,
	"scalar": KW_scalar,
}

type hasPosition interface{
	position() scanner.Position
}

func Position(i interface{}) (scanner.Position,bool) {
	hp,_ := i.(hasPosition)
	if hp==nil { return scanner.Position{},false }
	return hp.position(),true
}

type arrayExpr interface{
	array()
}

type ELiteral struct {
	Scalar values.Scalar
	Pos scanner.Position
}
func (e *ELiteral) String() string  {
	if _,ok := e.Scalar.(values.ScString); ok { return fmt.Sprintf("#%q",e.Scalar) }
	return fmt.Sprint("#",e.Scalar)
}
func (e *ELiteral) position() scanner.Position { return e.Pos }

type EScalar struct{ // $..
	Name interface{} // string | expression
	Pos scanner.Position
}
func (e *EScalar) String() string  { return fmt.Sprint("$",e.Name) }
func (e *EScalar) position() scanner.Position { return e.Pos }

type EHashScalar struct{ // $..{...}
	Name interface{} // string | expression
	Index interface{} // string | expression
	Pos scanner.Position
}
func (e *EHashScalar) String() string  { return fmt.Sprint("$",e.Name,"{",e.Index,"}") }
func (e *EHashScalar) position() scanner.Position { return e.Pos }

type EArrayScalar struct{ // $..[...]
	Name interface{} // string | expression
	Index interface{} // expression
	Pos scanner.Position
}
func (e *EArrayScalar) String() string  { return fmt.Sprint("$",e.Name,"[",e.Index,"]") }
func (e *EArrayScalar) position() scanner.Position { return e.Pos }

type EUnop struct{
	Op string // operation
	A interface{} // operand
	Pos scanner.Position
}
func (e *EUnop) String() string  { return fmt.Sprint("(",e.Op," ",e.A,")") }
func (e *EUnop) position() scanner.Position { return e.Pos }

type EBinop struct{
	Op string // operation
	A,B interface{} // operands
	Pos scanner.Position
}
func (e *EBinop) String() string  { return fmt.Sprint("(",e.A," ",e.Op," ",e.B,")") }
func (e *EBinop) position() scanner.Position { return e.Pos }

type EMatchGlobal struct{
	A interface{} // operand
	Rx *regexp.Regexp // regexp
	Pos scanner.Position
}
func (e *EMatchGlobal) String() string  { return fmt.Sprint("(",e.A," =~ m/",e.Rx,"/g)") }
func (e *EMatchGlobal) position() scanner.Position { return e.Pos }
func (e *EMatchGlobal) array() {}

type EMatch struct{
	A interface{} // operand
	Rx *regexp.Regexp // regexp
	Pos scanner.Position
}
func (e *EMatch) String() string  { return fmt.Sprint("(",e.A," =~ m/",e.Rx,"/)") }
func (e *EMatch) position() scanner.Position { return e.Pos }

type EReplace struct{
	A interface{} // operand
	Rx *regexp.Regexp // regexp
	B interface{} // operand (replacement)
	Pos scanner.Position
}
func (e *EReplace) String() string  { return fmt.Sprint("(",e.A," =~ s/",e.Rx,"/ ",e.B,")") }
func (e *EReplace) position() scanner.Position { return e.Pos }


type EScAssign struct{
	A,B interface{} // A := B
	Pos scanner.Position
}
func (e *EScAssign) String() string  { return fmt.Sprint(e.A," = ",e.B) }
func (e *EScAssign) position() scanner.Position { return e.Pos }


type EBinopAssign struct{
	Op string
	A,B interface{} // A <op>= B
	Pos scanner.Position
}
func (e *EBinopAssign) String() string  { return fmt.Sprint(e.A," ",e.Op,"= ",e.B) }
func (e *EBinopAssign) position() scanner.Position { return e.Pos }

type EFromArray struct{
	Array interface{}
	Pos scanner.Position
}
func (e *EFromArray) String() string { return fmt.Sprint(e.Array) }
func (e *EFromArray) position() scanner.Position { return e.Pos }

type AArray struct{ // @..
	Name interface{} // string | expression
	Pos scanner.Position
}
func (e *AArray) String() string  { return fmt.Sprint("@",e.Name) }
func (e *AArray) position() scanner.Position { return e.Pos }
func (e *AArray) array() {}

type AHash struct{ // %..
	Name interface{} // string | expression
	Pos scanner.Position
}
func (e *AHash) String() string  { return fmt.Sprint("@",e.Name) }
func (e *AHash) position() scanner.Position { return e.Pos }
func (e *AHash) array() {}

type AArAssign struct{
	A,B interface{} // A := B
	Pos scanner.Position
}
func (e *AArAssign) String() string  { return fmt.Sprint(e.A," = ",e.B) }
func (e *AArAssign) position() scanner.Position { return e.Pos }
func (e *AArAssign) array() {}



type SMyVars struct{ // my $a,@b,%c ...
	Vars []interface{} // variables (as string)
	Pos scanner.Position
}

type SExpr struct{ // <expression>;
	Expr interface{}
	Pos scanner.Position
}

type SArray struct{ // <expression>;
	Expr interface{}
	Pos scanner.Position
}


type SPrint struct { // print <expr>;
	Expr interface{}
	Pos scanner.Position
}

type SBlock struct{ // { ... }
	Stmts []interface{}
	Pos scanner.Position
}

type SCond struct{
	Type string
	Cond, Body interface{}
	Pos scanner.Position
}

type SIfElse struct{
	Type string
	Cond, Body, Else interface{}
	Pos scanner.Position
}

