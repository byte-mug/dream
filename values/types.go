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


package values

import "strconv"
import "fmt"
import "unsafe"
import "sync"

type Type uint

const (
	T_Nil Type = iota
	T_Integer
	T_Float
	T_String
	T_Buffer
	T_Reference
)

type Scalar interface{
	Type() Type
	IsFloat() bool
	Integer() int64
	Float() float64
	IsBytes() bool
	String() string
	Bytes() []byte
	AppendTo(prefix []byte) []byte
	Less(s Scalar) bool
}

type ScalarSlot interface{
	Get() Scalar
	Set(s Scalar)
}

type ModuleRef struct{
	Module unsafe.Pointer // *vm.Module
}
type ClassLoaderRef struct{
	ClassLoader unsafe.Pointer // *vm.ClassLoader
}

type nullt int
var null Scalar = nullt(0)

func(nullt) Type() Type { return T_Nil }
func(nullt) IsFloat() bool { return false }
func(nullt) Integer() int64 { return 0 }
func(nullt) Float() float64 { return 0 }
func(nullt) IsBytes() bool { return false }
func(nullt) String() string { return "" }
func(nullt) Bytes() []byte { return nil }
func(nullt) AppendTo(prefix []byte) []byte { return prefix }
func(nullt) Less(s Scalar) bool { return false }
func Null() Scalar { return null }

type ScInt int64
var _ Scalar = ScInt(0)
func (ScInt) Type() Type { return T_Integer }
func (ScInt) IsFloat() bool { return false }
func (v ScInt) Integer() int64 { return int64(v) }
func (v ScInt) Float() float64 { return float64(v) }
func (ScInt) IsBytes() bool { return false }
func (v ScInt) String() string { return strconv.FormatInt(int64(v),10) }
func (v ScInt) Bytes() []byte { return strconv.AppendInt(make([]byte,0,20),int64(v),10) }
func (v ScInt) AppendTo(prefix []byte) []byte { return strconv.AppendInt(prefix,int64(v),10) }
func (v ScInt) Less(s Scalar) bool { return v < s.(ScInt) }

type ScFloat float64
var _ Scalar = ScFloat(0)
func (ScFloat) Type() Type { return T_Float }
func (ScFloat) IsFloat() bool { return true }
func (v ScFloat) Integer() int64 { return int64(v) }
func (v ScFloat) Float() float64 { return float64(v) }
func (ScFloat) IsBytes() bool { return false }
func (v ScFloat) String() string { return strconv.FormatFloat(float64(v),'f',-1,64) }
func (v ScFloat) Bytes() []byte { return strconv.AppendFloat(make([]byte,0,30),float64(v),'f',-1,64) }
func (v ScFloat) AppendTo(prefix []byte) []byte { return strconv.AppendFloat(prefix,float64(v),'f',-1,64) }
func (v ScFloat) Less(s Scalar) bool { return v < s.(ScFloat) }

type ScString string
var _ Scalar = ScString("")
func (ScString) Type() Type { return T_String }
func (ScString) IsFloat() bool { return true }
func (s ScString) Integer() int64 {
	r,_ := strconv.ParseInt(string(s),0,64)
	return r
}
func (s ScString) Float() float64 {
	r,_ := strconv.ParseFloat(string(s),64)
	return r
}
func (ScString) IsBytes() bool { return false }
func (s ScString) String() string { return string(s) }
func (s ScString) Bytes() []byte { return []byte(s) }
func (s ScString) AppendTo(prefix []byte) []byte { return append(prefix,s...) }
func (v ScString) Less(s Scalar) bool { return v < s.(ScString) }

type ScBuffer []byte
var _ Scalar = ScBuffer{}
func (ScBuffer) Type() Type { return T_Buffer }
func (ScBuffer) IsFloat() bool { return true }
func (s ScBuffer) Integer() int64 {
	r,_ := strconv.ParseInt(string(s),0,64)
	return r
}
func (s ScBuffer) Float() float64 {
	r,_ := strconv.ParseFloat(string(s),64)
	return r
}
func (ScBuffer) IsBytes() bool { return true }
func (s ScBuffer) String() string { return string(s) }
func (s ScBuffer) Bytes() []byte { return []byte(s) }
func (s ScBuffer) AppendTo(prefix []byte) []byte { return append(prefix,s...) }
func (v ScBuffer) Less(s Scalar) bool { return string(v) < string(s.(ScBuffer)) }

type ScReference struct{
	Refid uintptr
	//Lock sync.RWMutex
	
	Lock sync.Mutex
	
	Data interface{}
	
	Blessed interface{}
}
func AllocScReference() *ScReference {
	r := new(ScReference)
	r.Refid = uintptr(unsafe.Pointer(r))
	return r
}
func (r *ScReference) Type() Type { return T_Reference }
func (r *ScReference) IsFloat() bool { return false }
func (r *ScReference) Integer() int64 { return int64(r.Refid) }
func (r *ScReference) Float() float64 { return float64(int64(r.Refid)) }
func (*ScReference) IsBytes() bool { return false }
func (r *ScReference) String() string {
	var t,c string
	switch r.Data.(type) {
	case *AV: t = "ARRAY"
	case *HV: t = "HASH"
	case ModuleRef: t = "CLASS"
	case ClassLoaderRef: t = "CLASSLOADER"
	}
	if b := r.Blessed; b!=nil { c = fmt.Sprintf("@%v",b) }
	return fmt.Sprintf("%s(0x%x)%s",t,r.Refid,c)
}
func (r *ScReference) Bytes() []byte { return []byte(r.String()) }
func (r *ScReference) AppendTo(prefix []byte) []byte { return append(prefix,r.String()...) }
func (r *ScReference) Less(s Scalar) bool { return r.Refid < s.(*ScReference).Refid }


