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


import "github.com/byte-mug/dream/values"
import "github.com/byte-mug/dream/vm"
import "regexp"
import "fmt"

//type vm.InsOp func(ts *vm.ThreadState, ip *int, ln int)

func debug(reg int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) { fmt.Println(ts.RS.SRegs[reg]) }
}

func literal(v values.Scalar, reg int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) { ts.RS.SRegs[reg] = v }
}

type slotLoader func(ts *vm.ThreadState) values.ScalarSlot

func load_global(n string, reg int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		rs := ts.RS
		v,ok := rs.Proc.Parent.Scalars.Load(n)
		if ok {
			rs.SRegs[reg] = *(v.(*values.Scalar))
		} else {
			rs.SRegs[reg] = values.Null()
		}
	}
}
func store_global(n string, reg int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		rs := ts.RS
		v,ok := rs.Proc.Parent.Scalars.Load(n)
		if !ok {
			v,_ = rs.Proc.Parent.Scalars.LoadOrStore(n,new(values.Scalar))
		}
		*(v.(*values.Scalar)) = rs.SRegs[reg]
	}
}
func slot_global(n string) slotLoader {
	return func(ts *vm.ThreadState) values.ScalarSlot {
		rs := ts.RS
		v,ok := rs.Proc.Parent.Scalars.Load(n)
		if ok {
			return values.MakeScalarSlot(v.(*values.Scalar))
		} else {
			return values.NonSlot()
		}
	}
}
func slot_local(reg int) slotLoader {
	return func(ts *vm.ThreadState) values.ScalarSlot {
		return values.MakeScalarSlot(&ts.RS.SRegs[reg])
	}
}

func load_unref(r1,rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		sr := ts.RS.SRegs
		sr[rT] = *(sr[r1].(*values.ScReference).Data.(*values.Scalar))
	}
}
func store_unref(r1,rSrc int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		sr := ts.RS.SRegs
		*(sr[r1].(*values.ScReference).Data.(*values.Scalar)) = sr[rSrc]
	}
}
func slot_unref(r1 int) slotLoader {
	return func(ts *vm.ThreadState) values.ScalarSlot {
		sr := ts.RS.SRegs
		return values.MakeScalarSlot(sr[r1].(*values.ScReference).Data.(*values.Scalar))
	}
}

type arrayLoader func(ts *vm.ThreadState) *values.AV
type hashLoader func(ts *vm.ThreadState) *values.HV

func avglobal(n string, w bool) arrayLoader {
	if !w {
		bol := new(values.AV)
		return func(ts *vm.ThreadState) *values.AV {
			v,ok := ts.RS.Proc.Parent.Arrays.Load(n)
			if !ok { v = bol }
			return v.(*values.AV)
		}
	}
	return func(ts *vm.ThreadState) *values.AV {
		v,ok := ts.RS.Proc.Parent.Arrays.Load(n)
		if !ok { v,_ = ts.RS.Proc.Parent.Arrays.LoadOrStore(n,new(values.AV)) }
		return v.(*values.AV)
	}
}
func hvglobal(n string, w bool) hashLoader {
	if !w {
		bol := new(values.HV)
		return func(ts *vm.ThreadState) *values.HV {
			v,ok := ts.RS.Proc.Parent.Hashes.Load(n)
			if !ok { v = bol }
			return v.(*values.HV)
		}
	}
	return func(ts *vm.ThreadState) *values.HV {
		v,ok := ts.RS.Proc.Parent.Hashes.Load(n)
		if !ok { v,_ = ts.RS.Proc.Parent.Hashes.LoadOrStore(n,new(values.HV)) }
		return v.(*values.HV)
	}
}
func avlocal(reg int) arrayLoader {
	return func(ts *vm.ThreadState) *values.AV { return &(ts.RS.ARegs[reg]) }
}
func hvlocal(reg int) hashLoader {
	return func(ts *vm.ThreadState) *values.HV { return &(ts.RS.HRegs[reg]) }
}
func avunref(reg int) arrayLoader {
	return func(ts *vm.ThreadState) *values.AV { return ts.RS.SRegs[reg].(*values.ScReference).Data.(*values.AV) }
}
func hvunref(reg int) hashLoader {
	return func(ts *vm.ThreadState) *values.HV { return ts.RS.SRegs[reg].(*values.ScReference).Data.(*values.HV) }
}

func load_array(al arrayLoader, r1, rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		av := al(ts)
		scrg := ts.RS.SRegs
		v := av.Fetch(scrg[r1].Integer(),false)
		if v==nil { v = values.Null() }
		scrg[rT] = v
	}
}
func store_array(al arrayLoader, r1, rSrc int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		av := al(ts)
		scrg := ts.RS.SRegs
		*av.Store(scrg[r1].Integer()) = scrg[rSrc]
	}
}
func slot_array(al arrayLoader, r1 int) slotLoader {
	return func(ts *vm.ThreadState) values.ScalarSlot {
		av := al(ts)
		scrg := ts.RS.SRegs
		return av.StoreSlot(scrg[r1].Integer())
	}
}
func load_hash(hl hashLoader, r1, rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		hv := hl(ts)
		scrg := ts.RS.SRegs
		slot := hv.Get(scrg[r1])
		if slot!=nil {
			scrg[rT] = slot.Get()
		} else {
			scrg[rT] = values.Null()
		}
	}
}
func store_hash(hl hashLoader, r1, rSrc int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		hv := hl(ts)
		scrg := ts.RS.SRegs
		hv.Put(scrg[r1]).Set(scrg[rSrc])
	}
}
func slot_hash(hl hashLoader, r1 int) slotLoader {
	return func(ts *vm.ThreadState) values.ScalarSlot {
		hv := hl(ts)
		scrg := ts.RS.SRegs
		return hv.Put(scrg[r1])
	}
}

/*

*/

func scalar_move(r1,rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		sr := ts.RS.SRegs
		sr[rT] = sr[r1]
	}
}

type binop_t func(a,b values.Scalar) values.Scalar

var binop_map = map[string]binop_t {
	"+": values.Add,
	"-": values.Sub,
	"*": values.Mul,
	"/": values.Div,
	"%": values.Mod,
	".": values.Concat,
	
	"and": values.And,
	"or": values.Or,
	"gt": values.GT,
	">": values.GT,
	"lt": values.LT,
	"<": values.LT,
	"ge": values.GE,
	">=": values.GE,
	"le": values.LE,
	"<=": values.LE,
	"eq": values.EQ,
	"==": values.EQ,
	"ne": values.NE,
	"!=": values.NE,
}

func binop(op binop_t, r1, r2, rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		sr := ts.RS.SRegs
		sr[rT] = op(sr[r1],sr[r2])
	}
}
func binop_assign(op binop_t, sl slotLoader, r2, rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		sr := ts.RS.SRegs
		slot := sl(ts)
		sr[rT] = op(slot.Get(),sr[r2])
		slot.Set(sr[rT])
	}
}

type unop_t func(a values.Scalar) values.Scalar

var unop_map = map[string]unop_t {
	"+": values.UPlus,
	"-": values.UMinus,
	"!": values.UNot,
	"~": values.UBitInv,
}

func unop(op unop_t, r1, rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		sr := ts.RS.SRegs
		sr[rT] = op(sr[r1])
	}
}


func regex_match(rx *regexp.Regexp, r1, rT int, regs []int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		sr := ts.RS.SRegs
		vs := sr[r1]
		sr[rT] = values.Null()
		if vs.IsBytes() {
			res := rx.FindSubmatch(vs.Bytes())
			if len(res)==0 { return }
			for i,v := range res {
				sr[regs[i]] = values.ScBuffer(v)
			}
		} else {
			res := rx.FindStringSubmatch(vs.String())
			if len(res)==0 { return }
			for i,v := range res {
				sr[regs[i]] = values.ScString(v)
			}
		}
		sr[rT] = values.Bool2S(true)
	}
}

func regex_replace(rx *regexp.Regexp, r1, r2, rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		sr := ts.RS.SRegs
		vs := sr[r1]
		repl := sr[r2]
		if vs.IsBytes() {
			sr[rT] = values.ScBuffer(rx.ReplaceAll(vs.Bytes(),repl.Bytes()))
		} else {
			sr[rT] = values.ScString(rx.ReplaceAllString(vs.String(),repl.String()))
		}
	}
}

func noop(ts *vm.ThreadState, ip *int, ln int) {
}
func last(ts *vm.ThreadState, ip *int, ln int) {
	ts.Flags |= vm.TSF_Last
	*ip = ln
}
func next(ts *vm.ThreadState, ip *int, ln int) {
	*ip = ln
}
func loop(slice []vm.InsOp) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		for {
			ts.RunSlice(slice)
			if (ts.Flags & vm.TSF_Last)!=0 { break }
		}
		ts.Flags &= ^vm.TSF_Last
		if (ts.Flags & vm.TSF_Return)!=0 { *ip = ln }
	}
}

func jump(off int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		*ip += off
	}
}
func jump_if(off, cond int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		if ts.RS.SRegs[cond].Bool() {
			*ip += off
		}
	}
}

func jump_unless(off, cond int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		if !ts.RS.SRegs[cond].Bool() {
			*ip += off
		}
	}
}

