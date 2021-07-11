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

func module(name string, reg int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		v := ts.RS.Proc.GetCl().GetModule(name)
		ts.RS.SRegs[reg] = v
	}
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

func load_array_global(n string, reg int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		rs := ts.RS
		v,ok := rs.Proc.Parent.Arrays.Load(n)
		if ok {
			rs.ARegs[reg] = append(rs.ARegs[reg],*(v.(*values.AV))...)
		} else {
			rs.ARegs[reg] = rs.ARegs[reg][:0]
		}
	}
}
func store_array_global(n string, reg int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		rs := ts.RS
		v,ok := rs.Proc.Parent.Arrays.Load(n)
		if ok {
			av := v.(*values.AV)
			*av = append((*av)[:0],rs.ARegs[reg]...)
		}
	}
}
func load_array_unref(r1,rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		sr := ts.RS.SRegs
		ar := ts.RS.ARegs
		ar[rT] = append(ar[rT][:0],*(sr[r1].(*values.ScReference).Data.(*values.AV))...)
	}
}
func store_array_unref(r1,rSrc int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		sr := ts.RS.SRegs
		ar := ts.RS.ARegs
		av := sr[r1].(*values.ScReference).Data.(*values.AV)
		*av = append((*av)[:0],ar[rSrc]...)
	}
}
func load_array_args(rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		ar := ts.RS.ARegs
		ar[rT] = ts.Args
	}
}
func commit_array_args(rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		ar := ts.RS.ARegs
		ts.Args = ar[rT]
	}
}
func store_array_args(r1 int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		ar := ts.RS.ARegs
		ts.Args = append(ts.Args[:0],ar[r1]...)
	}
}
func move_array(r1,rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		ar := ts.RS.ARegs
		ar[rT] = append(ar[rT][:0],ar[r1]...)
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

// arrayLoader
func avargs(ts *vm.ThreadState) *values.AV { return &(ts.Args) }

func hash_to_array(hl hashLoader, rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		avp := hl(ts).ToAV()
		ts.RS.ARegs[rT] = *avp
	}
}

func hash_from_array(hl hashLoader, rSrc int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		hv := hl(ts)
		hv.Clear()
		hv.FromAV(&ts.RS.ARegs[rSrc])
	}
}

func hash_transfer(hs, ht hashLoader) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		hvs := hs(ts)
		hvt := hs(ts)
		hvt.Clear()
		hvt.FromHV(hvs)
	}
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
func length_array(al arrayLoader, rT int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		av := al(ts)
		scrg := ts.RS.SRegs
		scrg[rT] = values.ScInt(len(*av))
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

func scratch_clear(rs int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		ARS := ts.RS.ARegs
		ARS[rs] = ARS[rs][:0]
	}
}
func scratch_null(rs int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		ts.RS.ARegs[rs] = nil
	}
}
func scratch_init(rs, r1 int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		ARS := ts.RS.ARegs
		ARS[rs] = ARS[r1]
	}
}
func scratch_add_scalar(rs, r1 int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		ARS := ts.RS.ARegs
		v := ts.RS.SRegs[r1]
		ARS[rs] = append(ARS[rs],v)
	}
}
func scratch_add_array(rs, r1 int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		ARS := ts.RS.ARegs
		ARS[rs] = append(ARS[rs],ARS[r1]...)
	}
}

func scratch_shift_scalar(rs, r1 int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		ARS := ts.RS.ARegs
		v := values.Null()
		if len(ARS)>0 {
			v = ARS[rs][0]
			ARS[rs] = ARS[rs][1:]
		}
		ts.RS.SRegs[r1] = v
	}
}
func scratch_shift_array(rs, r1 int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		ARS := ts.RS.ARegs
		ARS[r1] = append(ARS[r1][:0],ARS[rs]...)
		ARS[rs] = nil
	}
}

func scratch_create_array_ref(rs, r1 int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		av := ts.RS.ARegs[rs]
		nav := make(values.AV,len(av))
		copy(nav,av)
		sv := values.AllocScReference()
		sv.Data = &nav
		ts.RS.SRegs[r1] = sv
	}
}
func scratch_create_hash_ref(rs, r1 int) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		nhv := new(values.HV)
		nhv.FromAV(&ts.RS.ARegs[rs])
		sv := values.AllocScReference()
		sv.Data = nhv
		ts.RS.SRegs[r1] = sv
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

