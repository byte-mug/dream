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

import "math"

func Add(a, b Scalar) Scalar {
	if a.IsFloat() || b.IsFloat() {
		return ScFloat(a.Float()+b.Float())
	} else {
		return ScInt(a.Integer()+b.Integer())
	}
}

func Sub(a, b Scalar) Scalar {
	if a.IsFloat() || b.IsFloat() {
		return ScFloat(a.Float()-b.Float())
	} else {
		return ScInt(a.Integer()-b.Integer())
	}
}

func Mul(a, b Scalar) Scalar {
	if a.IsFloat() || b.IsFloat() {
		return ScFloat(a.Float()*b.Float())
	} else {
		return ScInt(a.Integer()*b.Integer())
	}
}

func Div(a, b Scalar) Scalar {
	if a.IsFloat() || b.IsFloat() {
		return ScFloat(a.Float()/b.Float())
	} else {
		return ScInt(a.Integer()/b.Integer())
	}
}

func Mod(a, b Scalar) Scalar {
	if a.IsFloat() || b.IsFloat() {
		return ScFloat(math.Remainder(a.Float(),b.Float()))
	} else {
		return ScInt(a.Integer()%b.Integer())
	}
}

func Concat(a, b Scalar) Scalar {
	if !a.IsBytes() { return ScString(a.String()+b.String()) }
	return ScBuffer(b.AppendTo(a.Bytes()))
}



func ScalarLess(a, b Scalar) bool {
	at,bt := a.Type(),b.Type()
	if at==bt { return a.Less(b) }
	return at<bt
}
func ScalarComp(a, b Scalar) int {
	at,bt := a.Type(),b.Type()
	if at==bt {
		if a.Less(b) { return -1 }
		if b.Less(a) { return 1 }
		return 0
	}
	if at<bt { return -1 }
	return 1
}
func RawScalarComp(a, b interface{}) int { return ScalarComp(a.(Scalar),b.(Scalar)) }
