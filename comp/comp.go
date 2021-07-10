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


package comp

import "github.com/byte-mug/dream/astparser"
//import "github.com/byte-mug/dream/values"
import "github.com/byte-mug/dream/vm"
import "regexp"
import "fmt"

type intAlloc struct {
	defined map[string]bool
	named map[string]int
	temp map[int]bool
	count map[int]int
}
func (ia *intAlloc) getTemp() (int,bool) {
	if ia.count==nil { ia.count = make(map[int]int) }
	for r,ok := range ia.temp {
		if !ok { continue }
		if ia.count[r]!=0 { continue }
		ia.count[r]++
		return r,true
	}
	return 0,false
}
func (ia *intAlloc) setTemp(r int) {
	if ia.temp==nil { ia.temp = make(map[int]bool) }
	if ia.count==nil { ia.count = make(map[int]int) }
	ia.temp[r] = true
	ia.count[r] = 0
}
func (ia *intAlloc) add(r, c int) {
	if ia.count==nil { ia.count = make(map[int]int) }
	if ia.temp[r] { ia.count[r] += c }
}
func (ia *intAlloc) getDefined(s string) (r int,ok bool) {
	r,ok = ia.named[s]
	return
}
func (ia *intAlloc) setDefined(s string, r int) {
	if ia.named==nil { ia.named = make(map[string]int) }
	ia.named[s] = r
}
func (ia *intAlloc) doDefine(s string, sigil string) {
	if ia.defined==nil { ia.defined = make(map[string]bool) }
	if ia.defined[s] { panic("Varialbe already declared: "+sigil+s) }
	ia.defined[s] = true
}

type Alloc struct {
	RSM vm.RSMetrics
	mgmt [vm.RSM_NumberOf]intAlloc
}
func (a *Alloc) temp(t int) int {
	var r int
	var ok bool
	if r,ok = a.mgmt[t].getTemp(); ok {
		return r
	} else {
		r = a.RSM[t]
		a.RSM[t] = r+1
		a.mgmt[t].setTemp(r)
	}
	a.mgmt[t].add(r,1)
	return r
}
func (a *Alloc) add(t,r,c int) {
	a.mgmt[t].add(r,c)
}
func (a *Alloc) defined(t int,s string) (int,bool) {
	return a.mgmt[t].getDefined(s)
}
func (a *Alloc) define(t int,s string, sigil string) {
	if sigil!="" { a.mgmt[t].doDefine(s,sigil) }
	_,ok := a.mgmt[t].getDefined(s)
	if !ok {
		r := a.RSM[t]
		a.RSM[t] = r+1
		a.mgmt[t].setDefined(s,r)
	}
}


// Scalar Target Hint
type ScTH int
const (
	ScAny ScTH = -(1+iota)
	ScDiscard
)
func (s ScTH) DeferDiscard() ScTH {
	if s==ScDiscard { return ScAny }
	return s
}

// -------------------------------
func (a *Alloc) GetScTarget(sth ScTH) int {
	if sth<0 { return a.temp(vm.RSM_Scalar) }
	return int(sth)
}
func (a *Alloc) PutScTarget(sth ScTH, r int) {
	if r<0 { return }
	if sth == ScDiscard { a.add(vm.RSM_Scalar,r,-1) }
}
func (a *Alloc) GetScDefined(s string) (int,bool) {
	return a.defined(vm.RSM_Scalar,s)
}
func (a *Alloc) SetScDefine(s string) {
	a.define(vm.RSM_Scalar,s,"$")
}
// -------------------------------
func (a *Alloc) GetArTarget(sth ScTH) int {
	if sth<0 { return a.temp(vm.RSM_Array) }
	return int(sth)
}
func (a *Alloc) PutArTarget(sth ScTH, r int) {
	if r<0 { return }
	if sth == ScDiscard { a.add(vm.RSM_Array,r,-1) }
}
func (a *Alloc) GetArDefined(s string) (int,bool) {
	return a.defined(vm.RSM_Array,s)
}
func (a *Alloc) SetArDefine(s string) {
	a.define(vm.RSM_Array,s,"@")
}
// -------------------------------
func (a *Alloc) GetHsTarget(sth ScTH) int {
	if sth<0 { return a.temp(vm.RSM_Hash) }
	return int(sth)
}
func (a *Alloc) PutHsTarget(sth ScTH, r int) {
	if r<0 { return }
	if sth == ScDiscard { a.add(vm.RSM_Hash,r,-1) }
}
func (a *Alloc) GetHsDefined(s string) (int,bool) {
	return a.defined(vm.RSM_Hash,s)
}
func (a *Alloc) SetHsDefine(s string) {
	a.define(vm.RSM_Hash,s,"%")
}
// -------------------------------
func (a *Alloc) MyDefine(s string) {
	switch s[0] {
	case '$': a.SetScDefine(s[1:])
	case '@': a.SetArDefine(s[1:])
	case '%': a.SetHsDefine(s[1:])
	}
}
// -------------------------------
func compileArrayLoader(alloc *Alloc, name interface{}, w bool) (ops []vm.InsOp, al arrayLoader, reg int) {
	if str,ok := name.(string); ok {
		if areg,ok := alloc.GetArDefined(str); ok {
			return nil,avlocal(areg),-1
		}
		return nil,avglobal(str,w),-1
	}
	ops,reg = ScCompile(alloc,name,ScAny)
	al = avunref(reg)
	return
}
func compileHashLoader(alloc *Alloc, name interface{}, w bool) (ops []vm.InsOp, al hashLoader, reg int) {
	if str,ok := name.(string); ok {
		if areg,ok := alloc.GetArDefined(str); ok {
			return nil,hvlocal(areg),-1
		}
		return nil,hvglobal(str,w),-1
	}
	ops,reg = ScCompile(alloc,name,ScAny)
	al = hvunref(reg)
	return
}

func scTarget(alloc *Alloc, targ, src interface{}, sth ScTH) (ops []vm.InsOp,reg int) {
	switch t := targ.(type) {
	case *astparser.EScalar:
		if str,ok := t.Name.(string); ok {
			if r1,ok := alloc.GetScDefined(str); ok {
				if sth<0 {
					reg = r1
				} else {
					reg = alloc.GetScTarget(sth)
				}
				
				o3,r3 := ScCompile(alloc,src,ScTH(r1))
				ops = o3
				if r1!=r3 { ops = append(ops,scalar_move(r3,r1)) }
				if reg!=r3 { ops = append(ops,scalar_move(r3,reg)) }
				return
			}
			ops,reg = ScCompile(alloc,src,sth.DeferDiscard())
			ops = append(ops,store_global(str,reg))
			alloc.PutScTarget(sth,reg)
		} else {
			o1,r1 := ScCompile(alloc,t.Name,ScAny)
			
			ops,reg = ScCompile(alloc,src,sth.DeferDiscard())
			ops = append(o1,ops...)
			ops = append(ops,store_unref(r1,reg))
			alloc.PutScTarget(ScDiscard,r1)
			alloc.PutScTarget(sth,reg)
		}
	case *astparser.EHashScalar:
		o1,al,r1 := compileHashLoader(alloc,t.Name,true)
		o2,r2 := ScCompile(alloc,t.Index,ScAny)
		o1 = append(o1,o2...)
		ops,reg = ScCompile(alloc,src,sth.DeferDiscard())
		ops = append(o1,ops...)
		ops = append(ops,store_hash(al,r2,reg))
		alloc.PutScTarget(ScDiscard,r1)
		alloc.PutScTarget(ScDiscard,r2)
		alloc.PutScTarget(sth,reg)
	case *astparser.EArrayScalar:
		o1,al,r1 := compileArrayLoader(alloc,t.Name,true)
		o2,r2 := ScCompile(alloc,t.Index,ScAny)
		o1 = append(o1,o2...)
		reg = alloc.GetScTarget(sth)
		ops = append(o1,ops...)
		ops = append(ops,store_array(al,r2,reg))
		alloc.PutScTarget(ScDiscard,r1)
		alloc.PutScTarget(ScDiscard,r2)
	default:
		pos,ok := astparser.Position(targ)
		if ok {
			panic(fmt.Errorf("%v : Can't assign to %v",pos,targ))
		} else {
			panic(fmt.Errorf("Can't assign to %v",targ))
		}
	}
	
	return
}
func scUpdate(alloc *Alloc, ast interface{}, try bool) (ops []vm.InsOp, sl slotLoader,regs []int) {
	var reg int
	switch t := ast.(type) {
	case *astparser.EScalar:
		if str,ok := t.Name.(string); ok {
			if reg,ok = alloc.GetScDefined(str); ok {
				sl = slot_local(reg)
			} else {
				sl = slot_global(str)
			}
		} else {
			ops,reg = ScCompile(alloc,t.Name,ScAny)
			sl = slot_unref(reg)
			regs = []int{reg}
		}
	case *astparser.EHashScalar:
		var al hashLoader
		var r1 int
		ops,al,r1 = compileHashLoader(alloc,t.Name,false)
		o2,r2 := ScCompile(alloc,t.Index,ScAny)
		ops = append(ops,o2...)
		sl = slot_hash(al,r2)
		regs = []int{r1,r2}
	case *astparser.EArrayScalar:
		var al arrayLoader
		var r1 int
		ops,al,r1 = compileArrayLoader(alloc,t.Name,false)
		o2,r2 := ScCompile(alloc,t.Index,ScAny)
		ops = append(ops,o2...)
		sl = slot_array(al,r2)
		regs = []int{r1,r2}
	default:
		if try { return nil,nil,nil }
		pos,ok := astparser.Position(ast)
		if ok {
			panic(fmt.Errorf("%v : Can't slot-assign to %v",pos,ast))
		} else {
			panic(fmt.Errorf("Can't slot-assign to %v",ast))
		}
	}
	
	return
}
func allocNumbers(alloc *Alloc, rx *regexp.Regexp) (regs []int) {
	n := rx.NumSubexp()+1
	regs = make([]int,n)
	for i := 0; i<n; i++ {
		S := fmt.Sprint(i)
		alloc.define(vm.RSM_Scalar,S,"")
		regs[i],_ = alloc.GetScDefined(S)
	}
	return
}

func ScCompile(alloc *Alloc, ast interface{}, sth ScTH) (ops []vm.InsOp,reg int) {
	switch t := ast.(type) {
	case *astparser.ELiteral:
		reg = alloc.GetScTarget(sth)
		ops = append(ops,literal(t.Scalar,reg))
		alloc.PutScTarget(sth,reg)
	case *astparser.EScalar:
		if str,ok := t.Name.(string); ok {
			if reg,ok = alloc.GetScDefined(str); ok { return }
			reg = alloc.GetScTarget(sth)
			ops = append(ops,load_global(str,reg))
			alloc.PutScTarget(sth,reg)
		} else {
			o1,r1 := ScCompile(alloc,t.Name,ScAny)
			reg = alloc.GetScTarget(sth)
			ops = o1
			ops = append(ops,load_unref(r1,reg))
			alloc.PutScTarget(ScDiscard,r1)
			alloc.PutScTarget(sth,reg)
		}
	case *astparser.EHashScalar:
		var al hashLoader
		var r1 int
		ops,al,r1 = compileHashLoader(alloc,t.Name,false)
		o2,r2 := ScCompile(alloc,t.Index,ScAny)
		ops = append(ops,o2...)
		reg = alloc.GetScTarget(sth)
		ops = append(ops,load_hash(al,r2,reg))
		alloc.PutScTarget(ScDiscard,r1)
		alloc.PutScTarget(ScDiscard,r2)
	case *astparser.EArrayScalar:
		var al arrayLoader
		var r1 int
		ops,al,r1 = compileArrayLoader(alloc,t.Name,false)
		o2,r2 := ScCompile(alloc,t.Index,ScAny)
		ops = append(ops,o2...)
		reg = alloc.GetScTarget(sth)
		ops = append(ops,load_array(al,r2,reg))
		alloc.PutScTarget(ScDiscard,r1)
		alloc.PutScTarget(ScDiscard,r2)
	case *astparser.EUnop:
		op,ok := unop_map[t.Op]
		if !ok { panic(t.Pos.String()+" Unary Operation not supported: "+t.Op) }
		o1,r1 := ScCompile(alloc,t.A,ScAny)
		reg = alloc.GetScTarget(sth)
		ops = o1
		ops = append(ops,unop(op,r1,reg))
		alloc.PutScTarget(ScDiscard,r1)
		alloc.PutScTarget(sth,reg)
	case *astparser.EBinop:
		op,ok := binop_map[t.Op]
		if !ok { panic(t.Pos.String()+" Binary Operation not supported: "+t.Op) }
		o1,r1 := ScCompile(alloc,t.A,ScAny)
		o2,r2 := ScCompile(alloc,t.B,ScAny)
		reg = alloc.GetScTarget(sth)
		ops = append(o1,o2...)
		ops = append(ops,binop(op,r1,r2,reg))
		alloc.PutScTarget(ScDiscard,r1)
		alloc.PutScTarget(ScDiscard,r2)
		alloc.PutScTarget(sth,reg)
	case *astparser.EMatchGlobal:
		panic(fmt.Errorf("%v Unsupported: %v",t.Pos,ast))
	case *astparser.EMatch:
		o1,r1 := ScCompile(alloc,t.A,ScAny)
		regs := allocNumbers(alloc,t.Rx)
		reg = alloc.GetScTarget(sth)
		ops = o1
		ops = append(ops,regex_match(t.Rx,r1,reg,regs))
		alloc.PutScTarget(ScDiscard,r1)
		alloc.PutScTarget(sth,reg)
	case *astparser.EReplace:
		o1,r1 := ScCompile(alloc,t.A,ScAny)
		o2,r2 := ScCompile(alloc,t.B,ScAny)
		reg = alloc.GetScTarget(sth)
		ops = append(o1,o2...)
		ops = append(ops,regex_replace(t.Rx,r1,r2,reg))
		alloc.PutScTarget(ScDiscard,r1)
		alloc.PutScTarget(ScDiscard,r2)
		alloc.PutScTarget(sth,reg)
	case *astparser.EScAssign:
		return scTarget(alloc,t.A,t.B,sth)
	case *astparser.EBinopAssign:
		op,ok := binop_map[t.Op]
		if !ok { panic(t.Pos.String()+" Binary Operation not supported: "+t.Op) }
		o1,sl,regs := scUpdate(alloc,t.A,false)
		o2,r2 := ScCompile(alloc,t.B,ScAny)
		regs = append(regs,r2)
		ops = append(o1,o2...)
		reg = alloc.GetScTarget(sth)
		ops = append(ops,binop_assign(op,sl,r2,reg))
		for _,oreg := range regs { alloc.PutScTarget(ScDiscard,oreg) }
		alloc.PutScTarget(sth,reg)
	default:
		pos,ok := astparser.Position(ast)
		if ok {
			panic(fmt.Errorf("%v : Expression not supported : %v",pos,ast))
		} else {
			panic(fmt.Errorf("Expression not supported : %v",ast))
		}
	}
	
	return
}

func StmtCompile(alloc *Alloc, ast interface{}) (ops []vm.InsOp) {
	switch t := ast.(type) {
	case *astparser.SMyVars:
		for _,s := range t.Vars { alloc.MyDefine(s.(string)) }
	case *astparser.SExpr:
		ops,_ = ScCompile(alloc,t.Expr,ScDiscard)
	case *astparser.SBlock:
		for _,s := range t.Stmts {
			o := StmtCompile(alloc,s)
			ops = append(ops,o...)
		}
	case *astparser.SCond:
		o1,r1 := ScCompile(alloc,t.Cond,ScAny)
		alloc.PutScTarget(ScDiscard,r1)
		o2 := StmtCompile(alloc,t.Body)
		switch t.Type {
		case "if":
			ops = append(o1,jump_unless(len(o2),r1))
			ops = append(ops,o2...)
		case "unless":
			ops = append(o1,jump_if(len(o2),r1))
			ops = append(ops,o2...)
		case "while":
			slice := append(o1,jump_if(1,r1),last)
			slice = append(slice,o2...)
			ops = append(ops,loop(slice))
		}
		ops = append(ops,noop)
	case *astparser.SIfElse:
		o1,r1 := ScCompile(alloc,t.Cond,ScAny)
		alloc.PutScTarget(ScDiscard,r1)
		o2 := StmtCompile(alloc,t.Body)
		o3 := StmtCompile(alloc,t.Else)
		switch t.Type {
		case "if":
			ops = append(o1,jump_unless(len(o2)+1,r1))
			ops = append(ops,o2...)
		case "unless":
			ops = append(o1,jump_if(len(o2)+1,r1))
			ops = append(ops,o2...)
		}
		ops = append(ops,jump(len(o3)))
		ops = append(ops,o3...)
	case *astparser.SPrint:
		o1,r1 := ScCompile(alloc,t.Expr,ScAny)
		alloc.PutScTarget(ScDiscard,r1)
		ops = append(o1,debug(r1)) // TODO: replace debug
	default:
		pos,ok := astparser.Position(ast)
		if ok {
			panic(fmt.Errorf("%v : Statement not supported : %v",pos,ast))
		} else {
			panic(fmt.Errorf("Statement not supported : %v",ast))
		}
	}
	return
}

func TestFunc(ast interface{}) (res *vm.Procedure) {
	res = new(vm.Procedure)
	res.Parent = new(vm.Module)
	
	alloc := new(Alloc)
	
	res.Instrs = StmtCompile(alloc,ast)
	res.Mets = alloc.RSM
	
	return
}
