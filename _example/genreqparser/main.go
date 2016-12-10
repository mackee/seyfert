package main

import (
	"os"
	"path/filepath"

	"github.com/mackee/seyfert"
)

func main() {
	binds := seyfert.Binds{
		"PATH": "root",
	}
	fieldsSet := seyfert.FieldsSet{
		"RequestFields": seyfert.Fields{
			seyfert.Field{
				Name: "HogeID",
				Type: "string",
				Tag:  `schema:"name"`,
			},
		},
	}
	tmplPath, err := filepath.Abs("../tmpl/reqparser.tmpl.go")
	if err != nil {
		panic(err)
	}
	generatePath, err := filepath.Abs("../rootparser.gen.go")
	if err != nil {
		panic(err)
	}
	bs, err := seyfert.Render(tmplPath, generatePath, binds, fieldsSet)
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(bs)
}
