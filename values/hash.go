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

import "sync"

type HV struct{
	Map sync.Map
}

/*
Obtains a Raw key for use as map-key.
*/
func Hv_Key(s Scalar) interface{} {
	switch v := s.(type) {
	case nullt: return null
	case ScInt: return int64(v)
	case ScFloat: return float64(v)
	case ScString: return string(v)
	case ScBuffer: return string(v) // Should not happen.
	case *ScReference: return v.Refid
	}
	return s.String()
}

/*
Curates a Scalar before being used as key in a HV. To ensure reproducible results,
buffers, whichs content may change will be converted into strings before use.
*/
func Hv_Curate(s Scalar) Scalar {
	if s.IsBytes() { return ScString(s.String()) }
	return s
}

type slotHV [2]Scalar
var _ ScalarSlot = &slotHV{nil}
func (hv *slotHV) Get() Scalar { return hv[1] }
func (hv *slotHV) Set(s Scalar) { hv[1] = s }

func (hv *HV) peek(s Scalar) *slotHV {
	v,ok := hv.Map.Load(Hv_Key(s))
	if ok { return v.(*slotHV) }
	return nil
}
func (hv *HV) poke(s Scalar) *slotHV {
	k := Hv_Key(s)
	v,ok := hv.Map.Load(k)
	if !ok { v,_ = hv.Map.LoadOrStore(k,&slotHV{s,null}) }
	return v.(*slotHV)
}
func (hv *HV) Get(key Scalar) ScalarSlot {
	sh := hv.peek(key)
	if sh!=nil { return sh }
	return nil
}
func (hv *HV) Put(key Scalar) ScalarSlot {
	return hv.poke(key)
}
func (hv *HV) Delete(key Scalar) {
	k := Hv_Key(key)
	hv.Map.Delete(k)
}
func (hv *HV) ToAV() *AV {
	avd := make(AV,0,16)
	hv.Map.Range(func(k,v interface{}) bool {
		if cap(avd)==len(avd) {
			navd := make(AV,len(avd),len(avd)*2)
			copy(navd,avd)
			avd = navd
		}
		avd = append(avd,v.(*slotHV)[:]...)
		return true
	})
	av := new(AV)
	*av = avd
	return av
}
func (hv *HV) FromAV(av *AV) {
	var key Scalar
	slot := new(slotHV)
	for _,s := range *av {
		if s==nil { s = null } // Safety first.
		if key==nil { key = Hv_Curate(s); continue }
		k := Hv_Key(key)
		*slot = slotHV{key,s}
		v,ld := hv.Map.LoadOrStore(k,slot)
		if ld { // if loaded, overwrite the entry
			*(v.(*slotHV)) = *slot
		} else { // if not, the slot is exthausted and must be replaced.
			slot = new(slotHV)
		}
	}
}
func (hv *HV) Clear(){
	var res = make([]interface{},0,16)
	hv.Map.Range(func(key, value interface{}) bool{
		res = append(res,key)
		return true
	})
	for _,k := range res { hv.Map.Delete(k) }
}
func (hv *HV) FromHV(hv2 *HV) {
	hv2.Map.Range(func(key, value interface{}) bool{
		slot := new(slotHV)
		*slot = *(value.(*slotHV))
		hv.Map.Store(key,slot)
		return true
	})
}


func Hv_Index(ref, idx Scalar) Scalar {
	slot := ref.(*ScReference).Data.(*HV).Get(idx)
	if slot==nil { return null }
	return slot.Get()
}
func Hv_IndexSlot(ref, idx Scalar) ScalarSlot {
	return ref.(*ScReference).Data.(*HV).Put(idx)
}

