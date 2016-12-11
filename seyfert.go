package seyfert

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/refactor/rename"
)

type Options struct {
	Binds Binds
}

type Binds map[string]string

type FieldsSet map[string]Fields

type Fields []Field

func (fs Fields) String() string {
	fss := make([]string, 0, len(fs))
	for _, f := range fs {
		fss = append(fss, f.String())
	}

	return strings.Join(fss, "\n")
}

type Field struct {
	Name string
	Type string
	Tag  string
}

func (f Field) String() string {
	if f.Tag == "" {
		return f.Name + " " + f.Type
	}

	return f.Name + " " + f.Type + " `" + f.Tag + "`"
}

type renameParam struct {
	offset string
	to     string
}

func Render(from string, to string, binds Binds, fieldsSet FieldsSet, packageName string) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, from, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var rps []renameParam
	for _, decl := range f.Decls {
		_rps := genDeclReanmeParam(decl, binds, to)
		if len(_rps) > 0 {
			rps = append(rps, _rps...)
		}
	}

	tof, err := os.Create(to)
	if err != nil {
		return nil, err
	}
	defer tof.Close()

	f.Name.Name = packageName
	err = format.Node(tof, fset, f)
	if err != nil {
		return nil, err
	}
	tof.Close()

	for _, p := range rps {
		fmt.Println(p.offset, "->", p.to)
		if err := rename.Main(&build.Default, p.offset, "", p.to); err != nil {
			return nil, err
		}
	}

	err = replaceExpandAnnotation(to, fieldsSet)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func tryBindName(name string, binds Binds) (string, bool) {
	var ok bool
	for key, binded := range binds {
		if strings.HasPrefix(name, key+"_") {
			ok = true
			name = strings.Replace(name, key+"_", binded, 1)
		}
		if strings.Contains(name, "_"+key+"_") {
			ok = true
			name = strings.Replace(name, "_"+key+"_", strings.Title(binded), -1)
		}
	}

	return name, ok
}

func genDeclReanmeParam(decl ast.Decl, binds Binds, filename string) []renameParam {
	var rps []renameParam

	switch t := decl.(type) {
	case *ast.FuncDecl:
		if !strings.HasPrefix(t.Doc.Text(), "+seyfert") {
			return nil
		}
		rp := genFuncRenameParam(t, binds, filename)
		if rp != nil {
			rps = append(rps, *rp)
		}
	case *ast.GenDecl:
		if !strings.HasPrefix(t.Doc.Text(), "+seyfert") {
			return nil
		}
		for _, spec := range t.Specs {
			typeSpec, isType := spec.(*ast.TypeSpec)
			if !isType {
				return nil
			}
			_rps := genTypeRenameParams(typeSpec, binds, filename)
			if len(_rps) > 0 {
				rps = append(rps, _rps...)
			}
		}
	default:
		return nil
	}

	return rps
}

func genFuncRenameParam(t *ast.FuncDecl, binds Binds, filename string) *renameParam {
	if binded, ok := tryBindName(t.Name.Name, binds); ok {
		return &renameParam{
			offset: fmt.Sprintf("%s:#%d", filename, t.Pos()),
			to:     binded,
		}
	}
	return nil
}

func genTypeRenameParams(t *ast.TypeSpec, binds Binds, filename string) []renameParam {
	var rps []renameParam
	if binded, ok := tryBindName(t.Name.Name, binds); ok {
		rps = append(rps, renameParam{
			offset: fmt.Sprintf("%s:#%d", filename, t.Pos()),
			to:     binded,
		})
	}

	return rps
}

type replaceParam struct {
	pos      int
	end      int
	replaced []byte
}

type replaceParams []replaceParam

func (rp replaceParams) Len() int {
	return len(rp)
}

func (rp replaceParams) Less(i, j int) bool {
	if rp[i].pos > rp[j].pos {
		return false
	}
	return true
}

func (rp replaceParams) Swap(i, j int) {
	rp[j], rp[i] = rp[i], rp[j]
}

func replaceExpandAnnotation(filename string, fieldsSet FieldsSet) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	bf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil
	}

	var rps replaceParams
	cmap := ast.NewCommentMap(fset, f, f.Comments)
	for _, cgg := range cmap {
		for _, cg := range cgg {
			for _, c := range cg.List {
				if !strings.HasPrefix(c.Text, "//+expand") {
					continue
				}
				fieldsName := strings.TrimPrefix(c.Text, "//+expand ")
				fields, ok := fieldsSet[fieldsName]
				if ok {
					rps = append(rps, replaceParam{
						pos:      int(c.Pos()),
						end:      int(c.End()),
						replaced: []byte(fields.String()),
					})
				}
			}
		}
	}

	sort.Sort(rps)
	foffset := 0
	buf := bytes.NewBuffer(make([]byte, 0, len(bf)))
	for _, rp := range rps {
		buf.Write(bf[foffset : rp.pos-1])
		buf.Write(rp.replaced)
		foffset = rp.end
	}
	buf.Write(bf[foffset:])

	out, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	ioutil.WriteFile(filename, out, 0644)

	return nil
}
