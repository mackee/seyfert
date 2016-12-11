package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mackee/seyfert"
)

func main() {
	binds := seyfert.Binds{
		"PATH":      "Root",
		"ROUTEPATH": "/root",
	}
	fieldsSet := seyfert.FieldsSet{
		"RequestFields": seyfert.Fields{
			seyfert.Field{
				Name: "HogeID",
				Type: "int",
				Tag:  `schema:"hoge_id"`,
			},
			seyfert.Field{
				Name: "Page",
				Type: "int",
				Tag:  `schema:"page"`,
			},
		},
		"ResponseFields": seyfert.Fields{
			seyfert.Field{
				Name: "HogeID",
				Type: "int",
				Tag:  `json:"hoge_id"`,
			},
		},
	}
	tmpl, err := ioutil.ReadFile("./tmpl/reqparser.tmpl.go")
	if err != nil {
		panic(err)
	}
	generatePath, err := filepath.Abs("../reqparser/rootparser.gen.go")
	if err != nil {
		panic(err)
	}
	bs, err := seyfert.Render(tmpl, generatePath, binds, fieldsSet, "reqparser")
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(bs)
}
