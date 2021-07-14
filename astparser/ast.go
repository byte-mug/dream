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
	KW_sub
	KW_for
	KW_do
	KW_eval
	KW_package
	KW_max_
)

var Keywords = scanlist.TokenDict{
	"undef"  : KW_undef,
	"and"    : KW_and,
	"or"     : KW_or,
	"eq"     : KW_eq,
	"ne"     : KW_ne,
	"lt"     : KW_lt,
	"le"     : KW_le,
	"gt"     : KW_gt,
	"ge"     : KW_ge,
	"m"      : KW_m,
	"s"      : KW_s,
	"my"     : KW_my,
	"if"     : KW_if,
	"unless" : KW_unless,
	"while"  : KW_while,
	"else"   : KW_else,
	"sub"    : KW_sub,
	"for"    : KW_for,
	"do"     : KW_do,
	"eval"   : KW_eval,
	"package": KW_package,
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
type hybridExpr interface{
	IsHybrid()
}
type callExpr interface{
	isCall()
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
func (e *EFromArray) String() string { return fmt.Sprint("scalar ",e.Array) }
func (e *EFromArray) position() scanner.Position { return e.Pos }

type ECreateArray struct{ // [ ... ]
	Elems []interface{}
	Pos scanner.Position
}
func (e *ECreateArray) String() string { return fmt.Sprint("newref (ARRAY) ",e.Elems) }
func (e *ECreateArray) position() scanner.Position { return e.Pos }

type ECreateHash struct{ // { ... }
	Elems []interface{}
	Pos scanner.Position
}
func (e *ECreateHash) String() string { return fmt.Sprint("newref (HASH) ",e.Elems) }
func (e *ECreateHash) position() scanner.Position { return e.Pos }

type EExIfElse struct{
	Cond, Then, Else interface{} // $a ? $b : $c
	Pos scanner.Position
}
func (e *EExIfElse) String() string  { return fmt.Sprint(e.Cond," ? ",e.Then," : ",e.Else) }
func (e *EExIfElse) position() scanner.Position { return e.Pos }

type EModule struct{
	Name string
	Pos scanner.Position
}
func (e *EModule) String() string  { return fmt.Sprint("module ",e.Name) }
func (e *EModule) position() scanner.Position { return e.Pos }

type ESubCall struct{
	Name string
	Args []interface{}
	Pos scanner.Position
}
func (e *ESubCall) String() string  { return fmt.Sprint("call ",e.Name, e.Args) }
func (e *ESubCall) position() scanner.Position { return e.Pos }
func (e *ESubCall) IsHybrid() {}
func (e *ESubCall) isCall() {}

type EObjCall struct{
	Obj interface{}
	Name string
	Args []interface{}
	Pos scanner.Position
}
func (e *EObjCall) String() string  { return fmt.Sprint("call (",e.Obj,")->",e.Name, e.Args) }
func (e *EObjCall) position() scanner.Position { return e.Pos }
func (e *EObjCall) IsHybrid() {}
func (e *EObjCall) isCall() {}

type EGoFunction struct{
	Call interface{}
	Pos scanner.Position
}
func (e *EGoFunction) String() string  { return fmt.Sprint("go ",e.Call) }
func (e *EGoFunction) position() scanner.Position { return e.Pos }
func (e *EGoFunction) IsHybrid() {}


func ToScalarExpr(ast interface{}) interface{} {
	if _,ok := ast.(hybridExpr); ok { return ast }
	if _,ok := ast.(arrayExpr); ok {
		pos,_ := Position(ast)
		return &EFromArray{ast,pos}
	}
	return ast
}
func IsArrayExpr(ast interface{}) bool {
	_,ok := ast.(arrayExpr)
	return ok
}

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

func flatten_one_level(args []interface{}) []interface{} {
	i := len(args)
	doflat := false
	for _,arg := range args {
		if cc,ok := arg.(*AConcat); ok {
			doflat = true
			i += len(cc.Elems)-1
		}
	}
	if !doflat { return args }
	nargs := make([]interface{},0,i)
	for _,arg := range args {
		if cc,ok := arg.(*AConcat); ok {
			nargs = append(nargs,cc.Elems...)
		} else {
			nargs = append(nargs,arg)
		}
	}
	return nargs
}

type AConcat struct{
	Elems []interface{} // ($a,$b,$c,...)
	Pos scanner.Position
}
func (e *AConcat) String() string  { return fmt.Sprint(e.Elems) }
func (e *AConcat) position() scanner.Position { return e.Pos }
func (e *AConcat) array() {}


type AExIfElse struct{
	Cond, Then, Else interface{} // $a ? @b : @c
	Pos scanner.Position
}
func (e *AExIfElse) String() string  { return fmt.Sprint(e.Cond," ? ",e.Then," : ",e.Else) }
func (e *AExIfElse) position() scanner.Position { return e.Pos }
func (e *AExIfElse) array() {}

func ToArrayExpr(ast interface{}) interface{} {
	if _,ok := ast.(hybridExpr); ok { return ast }
	if _,ok := ast.(arrayExpr); !ok {
		pos,_ := Position(ast)
		return &AConcat{[]interface{}{ast},pos}
	}
	return ast
}

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
type SNoop struct{
	Pos scanner.Position
}
type SFor struct{ // for $a (@b) {...}
	Var string
	Src, Body interface{}
	Pos scanner.Position
}
type SEval struct{
	Body interface{}
	Pos scanner.Position
}

type SLoopJump struct{
	Op string // next | last
	Pos scanner.Position
}

type MDPackage struct{
	Name string
	Pos scanner.Position
}

type MDSub struct{
	Name string
	Body interface{}
	Pos scanner.Position
}


type Module struct{
	Main *MDSub
	Subs []*MDSub
}

