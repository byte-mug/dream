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


package binding

//import "github.com/byte-mug/dream/values"
import "github.com/byte-mug/dream/vm"
import "reflect"

type simpleFunc func(ts *vm.ThreadState)

func wrap_simpleFunc(fu simpleFunc) vm.InsOp {
	return func(ts *vm.ThreadState, ip *int, ln int) {
		fu(ts)
	}
}

func convertFunc(fu interface{}) vm.InsOp {
	switch t := fu.(type) {
	case func(ts *vm.ThreadState, ip *int, ln int):
		return vm.InsOp(t)
	case func(ts *vm.ThreadState):
		return wrap_simpleFunc(simpleFunc(t))
	}
	return nil
}

func convertFuncObj(mod *vm.Module, fu interface{}) *vm.Procedure {
	op := convertFunc(fu)
	if op==nil { return nil }
	
	p := new(vm.Procedure)
	p.Parent = mod
	p.Instrs = []vm.InsOp{op}
	return p
}

func convertTypeObject(cl *vm.ClassLoader, name string,v reflect.Value) *vm.Module {
	md := &vm.Module{Parent: cl, Name: name}
	
	md.Main = &vm.Procedure{Parent:md}
	
	t := v.Type()
	for i,n := 0,t.NumMethod(); i<n; i++ {
		p := convertFuncObj(md,v.Method(i).Interface())
		if p==nil { continue }
		name := t.Method(i).Name
		md.Procedures.Store(name,p)
	}
	
	return md
}

func CreateBindingModule(cl *vm.ClassLoader, name string,v interface{}) *vm.Module {
	return convertTypeObject(cl,name,reflect.ValueOf(v))
}

