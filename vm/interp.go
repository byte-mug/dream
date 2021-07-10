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


package vm

import "github.com/byte-mug/dream/values"
import "github.com/byte-mug/dream/allocrefl"
import "sync"

/*
This structure contains everything that is supposed to be global and thread-local.
*/
type ThreadState struct {
	RS *RegisterSet
	
	Flags uint
}

const (
	TSF_Last uint = 1<<iota
	TSF_Return
)

type InsOp func(ts *ThreadState, ip *int, ln int)

var sRegs = allocrefl.Allocator{ Sample: []values.Scalar{} }
var aRegs = allocrefl.Allocator{ Sample: []values.AV{} }
//var hRegs = allocrefl.Allocator{ Sample: []values.HV{} }

func sInit(a []values.Scalar) {
	null := values.Null()
	for i := range a{
		a[i] = null
	}
}
func sWipe(a []values.Scalar) {
	for i := range a{
		a[i] = nil
	}
}
func aWipe(a []values.AV) {
	for i := range a{
		a[i] = nil
	}
}
//func hWipe(a []values.HV) {}


const (
	RSM_Scalar = iota
	RSM_Array
	RSM_Hash
	RSM_NumberOf
)

type RSMetrics [RSM_NumberOf]int
func (rsm RSMetrics) Alloc() *RegisterSet {
	rs := new(RegisterSet)
	rs.SRegs = sRegs.Alloc(rsm[RSM_Scalar]).([]values.Scalar)[:rsm[RSM_Scalar]]
	rs.ARegs = aRegs.Alloc(rsm[RSM_Array]).([]values.AV)[:rsm[RSM_Array]]
	//rs.HRegs = hRegs.Alloc(rsm[2]).([]values.HV)[:rsm[2]]
	rs.HRegs = make([]values.HV,rsm[RSM_Hash])
	sInit(rs.SRegs)
	return rs
}
type RegisterSet struct {
	SRegs []values.Scalar
	ARegs []values.AV
	HRegs []values.HV
	
	Proc *Procedure
}

func (rs *RegisterSet) Sproc(p *Procedure) *RegisterSet {
	rs.Proc = p
	return rs
}

func (rs *RegisterSet) Set(ts *ThreadState) (old *RegisterSet) {
	old = ts.RS
	ts.RS = rs
	return
}
func (rs *RegisterSet) SetDispose(ts *ThreadState) {
	old := rs.Set(ts)
	sWipe(old.SRegs)
	aWipe(old.ARegs)
	sRegs.FreeRaw(len(old.SRegs),old.SRegs)
	aRegs.FreeRaw(len(old.ARegs),old.ARegs)
	//hRegs.FreeRaw(len(old.HRegs),old.HRegs)
	*old = RegisterSet{}
}

type ClassLoader struct{
	Modules sync.Map // map[string]*Module
}

type Module struct{
	Parent *ClassLoader
	Procedures sync.Map // map[string]*Procedure
	
	Scalars sync.Map // map[string]*values.Scalar
	Arrays sync.Map // map[string]*values.AV
	Hashes sync.Map // map[string]*values.HV
}

type Procedure struct{
	Parent *Module
	Mets RSMetrics
	Instrs []InsOp
}

func (p *Procedure) Exec(ts *ThreadState) {
	defer p.Mets.Alloc().Sproc(p).Set(ts).SetDispose(ts)
	slice := p.Instrs
	i,n := 0,len(slice)
	for i<n {
		f := slice[i]
		i++
		f(ts,&i,n)
	}
	ts.Flags &= ^TSF_Return
}

func (ts *ThreadState) RunSlice(slice []InsOp) {
	i,n := 0,len(slice)
	for i<n {
		f := slice[i]
		i++
		f(ts,&i,n)
	}
}

