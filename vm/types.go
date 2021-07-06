/*
Copyright (c) 2021 byte-mug

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

type Type uint

const (
	T_Nil Type = iota
	T_Integer
	T_Float
	T_String
	T_Buffer
)

type Scalar interface{
	Type() Type
}

type ScalarNumber interface{
	Scalar
	IsFloat() bool
	Integer() int64
	Float() float64
}

type nullt int
func(nullt) Type() Type { return T_Nil }
var null Scalar = nullt(0)

func Null() Scalar { return null }

type ScInt int64
func (ScInt) Type() Type { return T_Integer }
func (ScInt) IsFloat() bool { return false }
func (v ScInt) Integer() int64 { return int64(v) }
func (v ScInt) Float() float64 { return float64(v) }


type ScFloat float64
func (ScFloat) Type() Type { return T_Float }
func (ScFloat) IsFloat() bool { return true }
func (v ScFloat) Integer() int64 { return int64(v) }
func (v ScFloat) Float() float64 { return float64(v) }

type ScString string
func (ScString) Type() Type { return T_String }

type ScBuffer []byte
func (ScBuffer) Type() Type { return T_Buffer }



