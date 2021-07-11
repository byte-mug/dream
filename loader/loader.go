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


package loader

import (
	"github.com/byte-mug/semiparse/scanlist"
	"github.com/byte-mug/semiparse/parser"
	
	"github.com/byte-mug/dream/astparser"
	"github.com/byte-mug/dream/vm"
	"github.com/byte-mug/dream/comp"
	"fmt"
)
import (
	"io"
	"os"
	fpath "path/filepath"
	"strings"
	
)

type GenericLoader struct{
	Parser parser.Parser
}
func CreateGenericLoader() *GenericLoader {
	gl := new(GenericLoader)
	gl.Parser.Construct()
	astparser.Register(&gl.Parser)
	return gl
}
func (gl *GenericLoader) load(cl *vm.ClassLoader, r io.Reader, name, fn string) *vm.Module {
	var bs scanlist.BaseScanner
	bs.Init(r)
	bs.Filename = fn
	bs.Dict = astparser.Keywords
	res := gl.Parser.Match("Module",bs.Next())
	if !res.Ok() { panic(fmt.Sprint(res.Pos," : ",res.Data)) }
	sm := res.Data.(*astparser.Module)
	return comp.ModCompile(cl,name,sm)
}
func (gl *GenericLoader) LoadModuleFrom(paths []string, cl *vm.ClassLoader, name string) *vm.Module {
	s := string([]byte{os.PathSeparator})
	fp := strings.Replace(name,"::",s,-1)+".dm"
	for _,pth := range paths {
		absp := fpath.Join(pth,fp)
		fobj,err := os.Open(absp)
		if err!=nil { continue }
		defer fobj.Close()
		_,fn := fpath.Split(absp)
		return gl.load(cl,fobj,name,fn)
	}
	return nil
	panic("...")
}

type SpecificLoader struct{
	GL *GenericLoader
	Paths []string
}
var _ vm.ClassLoaderSpi = (*SpecificLoader)(nil)
func (sl *SpecificLoader) LoadModule(cl *vm.ClassLoader, name string) *vm.Module {
	return sl.GL.LoadModuleFrom(sl.Paths,cl,name)
}

