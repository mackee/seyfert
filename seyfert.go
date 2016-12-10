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

	pp "gopkg.in/pp.v2"
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
		switch t := decl.(type) {
		case *ast.FuncDecl:
			if !strings.HasPrefix(t.Doc.Text(), "+seyfert") {
				continue
			}
			//bindFunc(t, binds)
			if binded, ok := tryBindName(t.Name.Name, binds); ok {
				rps = append(rps, renameParam{
					offset: fmt.Sprintf("%s:#%d", to, t.Pos()),
					to:     binded,
				})
			}
		case *ast.GenDecl:
			if !strings.HasPrefix(t.Doc.Text(), "+seyfert") {
				continue
			}
			for _, spec := range t.Specs {
				typeSpec, isType := spec.(*ast.TypeSpec)
				if !isType {
					continue
				}
				if binded, ok := tryBindName(typeSpec.Name.Name, binds); ok {
					rps = append(rps, renameParam{
						offset: fmt.Sprintf("%s:#%d", to, typeSpec.Pos()),
						to:     binded,
					})
				}
			}
		default:
			continue
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

func bindType(t *ast.TypeSpec, binds Binds) {
	if binded, ok := tryBindName(t.Name.Name, binds); ok {
		t.Name.Name = binded
	}
}

func bindFunc(t *ast.FuncDecl, binds Binds) {
	if binded, ok := tryBindName(t.Name.Name, binds); ok {
		t.Name.Name = binded
	}

	if t.Recv != nil {
		pp.Println(t.Recv)
		for _, field := range t.Recv.List {
			ident, isIdent := field.Type.(*ast.Ident)
			if !isIdent {
				continue
			}
			if binded, ok := tryBindName(ident.Name, binds); ok {
				ident.Name = binded
			}
		}
	}

	if t.Type.Params.NumFields() > 0 {
		for _, field := range t.Type.Params.List {
			ident, isIdent := field.Type.(*ast.Ident)
			if !isIdent {
				continue
			}
			if binded, ok := tryBindName(ident.Name, binds); ok {
				ident.Name = binded
			}
		}
	}
}
