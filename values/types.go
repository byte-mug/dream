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

type Type uint

const (
	T_Nil Type = iota
	T_Integer
	T_Float
	T_String
	T_Buffer
	T_Reference
	T_Module
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
	Bool() bool
}

type ScalarSlot interface{
	Get() Scalar
	Set(s Scalar)
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
func(nullt) Bool() bool { return false }
func Null() Scalar { return null }

type nonSlot_t int
var nonSlot ScalarSlot = nonSlot_t(0)
func (nonSlot_t) Get() Scalar { return null }
func (nonSlot_t) Set(s Scalar) { }
func NonSlot() ScalarSlot { return nonSlot }


type ScInt int64
var strue Scalar = ScInt(1)
func (ScInt) Type() Type { return T_Integer }
func (ScInt) IsFloat() bool { return false }
func (v ScInt) Integer() int64 { return int64(v) }
func (v ScInt) Float() float64 { return float64(v) }
func (ScInt) IsBytes() bool { return false }
func (v ScInt) String() string { return strconv.FormatInt(int64(v),10) }
func (v ScInt) Bytes() []byte { return strconv.AppendInt(make([]byte,0,20),int64(v),10) }
func (v ScInt) AppendTo(prefix []byte) []byte { return strconv.AppendInt(prefix,int64(v),10) }
func (v ScInt) Less(s Scalar) bool { return v < s.(ScInt) }
func (v ScInt) Bool() bool { return v!=0 }

func Bool2S(b bool) Scalar {
	if b {
		return strue
	} else {
		return null
	}
}

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
func (v ScFloat) Bool() bool { return v!=0 }

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
func (s ScString) Bool() bool {
	if len(s)==0 { return false }
	r,e := strconv.ParseInt(string(s),0,64)
	return r!=0 || e!=nil
}

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
func (s ScBuffer) Bool() bool {
	if len(s)==0 { return false }
	r,e := strconv.ParseInt(string(s),0,64)
	return r!=0 || e!=nil
}

type ScReference struct{
	Refid uintptr
	
	Data interface{}
	
	Blessed *ScModule
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
	case *Scalar: t = "SCALAR"
	case *AV: t = "ARRAY"
	case *HV: t = "HASH"
	case ClassLoaderRef: t = "CLASSLOADER"
	}
	if b := r.Blessed; b!=nil { c = fmt.Sprintf("%v=",b) }
	return fmt.Sprintf("%s%s(0x%x)",c,t,r.Refid)
}
func (r *ScReference) Bytes() []byte { return []byte(r.String()) }
func (r *ScReference) AppendTo(prefix []byte) []byte { return append(prefix,r.String()...) }
func (r *ScReference) Less(s Scalar) bool { return r.Refid < s.(*ScReference).Refid }
func (*ScReference) Bool() bool { return true }


type ScModule struct{
	Name string
	DisplayName string
	Clid uintptr
	ClassLoader interface{} // *vm.ClassLoader
	ModuleObject interface{}// *vm.Module or nil
}

func (r *ScModule) Type() Type { return T_Module }
func (r *ScModule) IsFloat() bool { return false }
func (r *ScModule) Integer() int64 { return 1 }
func (r *ScModule) Float() float64 { return 1 }
func (*ScModule) IsBytes() bool { return false }
func (r *ScModule) String() string { return r.DisplayName }
func (r *ScModule) Bytes() []byte { return []byte(r.String()) }
func (r *ScModule) AppendTo(prefix []byte) []byte { return append(prefix,r.String()...) }
func (r *ScModule) Less(s Scalar) bool {
	o := s.(*ScModule)
	if r.Clid != o.Clid { return r.Clid < o.Clid }
	return r.Name < o.Name
}
func (*ScModule) Bool() bool { return true }

func AllocNewScModule(n string,clid uintptr,cl interface{}) *ScModule {
	dn := n
	if clid!=0 { dn += fmt.Sprintf("<0x%x>",clid) }
	return &ScModule{n,dn,clid,cl,nil}
}

func GetScModule(sc Scalar) *ScModule {
	switch t := sc.(type) {
	case *ScModule:
		return t
	case *ScReference:
		return t.Blessed
	}
	return nil
}

