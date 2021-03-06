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

	"github.com/pkg/errors"
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

func Render(tmpl []byte, to string, binds Binds, fieldsSet FieldsSet, packageName string) error {
	err := ioutil.WriteFile(to, tmpl, 0644)
	if err != nil {
		return errors.Wrap(err, "cannot write destination file error")
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, to, nil, parser.ParseComments)
	if err != nil {
		return errors.Wrap(err, "fail parsing error")
	}

	var rps []renameParam
	for _, decl := range f.Decls {
		_rps := genDeclReanmeParam(decl, binds, to)
		if len(_rps) > 0 {
			rps = append(rps, _rps...)
		}
	}

	tof, err := os.OpenFile(to, os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "fail create destination file error")
	}
	defer tof.Close()

	f.Name.Name = packageName
	err = format.Node(tof, fset, f)
	if err != nil {
		return errors.Wrap(err, "fail write template stage file error")
	}
	tof.Close()

	for _, p := range rps {
		fmt.Println(p.offset, "->", p.to)
		if err := rename.Main(&build.Default, p.offset, "", p.to); err != nil {
			return errors.Wrap(err, "fail rename error")
		}
	}

	err = replaceExpandAnnotation(to, fieldsSet)
	if err != nil {
		return errors.Wrap(err, "expand annotation error")
	}

	err = replaceLiteralInFunc(to, binds)
	if err != nil {
		return errors.Wrap(err, "replace literal error")
	}

	return nil
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
			name = strings.Replace(name, "_"+key+"_", binded, -1)
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
	var rp *renameParam
	if binded, ok := tryBindName(t.Name.Name, binds); ok {
		rp = &renameParam{
			offset: fmt.Sprintf("%s:#%d", filename, t.Pos()),
			to:     binded,
		}
	}

	return rp
}

func replaceLiteralInFunc(filename string, binds Binds) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return errors.Wrap(err, "cannot pasing error")
	}

	for _, decl := range f.Decls {
		funcDecl, isFunc := decl.(*ast.FuncDecl)
		if !isFunc {
			continue
		}
		if !strings.HasPrefix(funcDecl.Doc.Text(), "+seyfert") {
			continue
		}

		for _, stmt := range funcDecl.Body.List {
			ast.Inspect(stmt, func(node ast.Node) bool {
				lit, ok := node.(*ast.BasicLit)
				if ok {
					if binded, ok := tryBindName(lit.Value, binds); ok {
						lit.Value = binded
					}
				}
				return true
			})
		}
	}

	tof, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrap(err, "cannot open error")
	}
	defer tof.Close()

	err = format.Node(tof, fset, f)
	if err != nil {
		return errors.Wrap(err, "cannot write node error")
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
		return errors.Wrap(err, "cannot parsing error")
	}

	bf, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "cannnot read file error")
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
		return errors.Wrap(err, "fail node to source error")
	}

	ioutil.WriteFile(filename, out, 0644)

	return nil
}
