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


package allocrefl

import "reflect"
import "sync"

const mSIZE = 10
func findSize(siz int) uint {
	for i := uint(0); i<mSIZE; i++ {
		if siz <= 1<<i { return i }
	}
	return mSIZE
}

type Allocator struct{
	Sample interface{}
	tSample reflect.Type
	
	pools [mSIZE]sync.Pool
}
func (a *Allocator) init() {
	if a.Sample==nil { panic("No specimen") }
	if a.tSample==nil { a.tSample = reflect.TypeOf(a.Sample) }
}
func (a *Allocator) mak(size int) interface{} {
	return reflect.MakeSlice(a.tSample,size,size).Interface()
}
func (a *Allocator) Alloc(size int) interface{} {
	i := findSize(size)
	if i==mSIZE { return a.mak(size) }
	v := a.pools[i].Get()
	if v==nil { v = a.mak(1<<i) }
	return v
}
func (a *Allocator) FreeRaw(size int, arr interface{}) {
	i := findSize(size)
	if i<mSIZE { a.pools[i].Put(arr) }
}


