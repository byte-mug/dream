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

type AV []Scalar

type slotAV [1]*Scalar
var _ ScalarSlot = &slotAV{nil}
func (av *slotAV) Get() Scalar { return *(av[0]) }
func (av *slotAV) Set(s Scalar) { *(av[0]) = s }

//Returns the highest index of the array (such as $#array).
func (av *AV) Len() int { return len(*av) }

/*
Retrieves the scalar from the given index. If unset is true, it replaces the existing value (at that location) with an null.
*/
func (av *AV) Fetch(i int64, unset bool) Scalar {
	avd := *av
	if int64(len(avd))<=i { return nil }
	r := avd[i]
	if r!=nil && unset { avd[i] = null }
	return r
}

/*
Retrieves the scalar from the given index. If unset is true, it replaces the existing value (at that location) with an null.
Note that FetchUp returns *Scalar, not Scalar. This way you can update the slot as you wish.
*/
func (av *AV) FetchUp(i int64, unset bool) *Scalar {
	avd := *av
	if int64(len(avd))<=i { return nil }
	r := &avd[i]
	if *r!=nil && unset { *r = null }
	return r
}

/*
Retrieves the scalar from the given index. If unset is true, it replaces the existing value (at that location) with an null.
Note that FetchSlot returns ScalarSlot, not Scalar. This way you can update the slot as you wish.
*/
func (av *AV) FetchSlot(i int64, unset bool) ScalarSlot {
	p := av.FetchUp(i,unset)
	if p!=nil { return &slotAV{p} }
	return nil
}

func (av *AV) Store(i int64) *Scalar {
	avd := *av
	o := int64(len(avd))
	if o>i { return &avd[i] }
	if int64(cap(avd))>i {
		avd = avd[:i+1]
	} else {
		avd = make([]Scalar,i+1)
		copy(avd,*av)
	}
	for ;o<i; o++ { avd[o] = null }
	*av = avd
	return &avd[i]
}
func (av *AV) StoreSlot(i int64) ScalarSlot {
	p := av.Store(i)
	if p==nil { return nil }
	*p = null
	return &slotAV{p}
	return nil
}

func (av *AV) Push(s Scalar) { *av = append(*av,s) }
func (av *AV) Pop(s Scalar) Scalar {
	avd := *av
	r := avd[len(avd)-1]
	*av = avd[:len(avd)-1]
	return r
}

func Av_Index(ref, idx Scalar) Scalar {
	return ref.(*ScReference).Data.(*AV).Fetch(idx.Integer(),false)
}
func Av_IndexSlot(ref, idx Scalar) ScalarSlot {
	av := ref.(*ScReference).Data.(*AV)
	i := idx.Integer()
	sl := av.FetchSlot(i,false)
	if sl==nil { sl = av.StoreSlot(i) }
	return sl
}


