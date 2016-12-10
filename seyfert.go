package seyfert

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strings"

	"golang.org/x/tools/refactor/rename"
)

type Options struct {
	Binds Binds
}

type Binds map[string]string

type FieldsSet map[string]Fields

type Fields []Field

type Field struct {
	Name string
	Type string
	Tag  string
}

type renameParam struct {
	offset string
	to     string
}

func Render(from string, to string, binds Binds, fieldsSet FieldsSet) ([]byte, error) {
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

	fromf, err := os.Open(from)
	if err != nil {
		return nil, err
	}
	defer fromf.Close()

	_, err = io.Copy(tof, fromf)
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
