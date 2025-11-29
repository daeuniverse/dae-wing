/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package main

import (
	"fmt"
	daeConfig "github.com/daeuniverse/dae/config"
	"github.com/sirupsen/logrus"
	"github.com/stoewer/go-strcase"
	"os"
	"reflect"
	"strings"
	"time"
)

type builder struct {
	sb strings.Builder
}

func (b *builder) WriteLine(depth int, line string) {
	b.sb.WriteString(strings.Repeat("\t", depth))
	b.sb.WriteString(line)
	b.sb.WriteString("\n")
}

func (b *builder) WriteFunc(fieldName string, name string, retTyp string, cast bool) {
	b.WriteLine(0, fmt.Sprintf("func (r *Resolver) %v() %v {", strcase.UpperCamelCase(name), retTyp))
	if cast {
		b.WriteLine(1, fmt.Sprintf("return %v(r.Global.%v)", retTyp, fieldName))
	} else {
		b.WriteLine(1, fmt.Sprintf("return r.Global.%v", fieldName))
	}
	b.WriteLine(0, fmt.Sprintf("}\n"))
}

func (b *builder) WriteMarshalFunc(fieldName string, name string, retTyp string) {
	b.WriteLine(0, fmt.Sprintf("func (r *Resolver) %v() %v {", strcase.UpperCamelCase(name), retTyp))
	b.WriteLine(1, fmt.Sprintf("var tmp %v", retTyp))
	b.WriteLine(1, fmt.Sprintf("_ = tmp.UnmarshalGraphQL(r.Global.%v)", fieldName))
	b.WriteLine(1, fmt.Sprintf("return tmp"))
	b.WriteLine(0, fmt.Sprintf("}\n"))
}

func (b *builder) Build() (string, error) {

	v := reflect.ValueOf(daeConfig.Global{})
	t := v.Type()
	b.WriteLine(0, "// Generated code; DO NOT EDIT.\n")
	b.WriteLine(0, "package global\n")
	b.WriteLine(0, fmt.Sprintf(`import "%v"`, t.PkgPath()))
	b.WriteLine(0, fmt.Sprintf(`import "github.com/daeuniverse/dae-wing/graphql/scalar"`))
	b.WriteLine(0, `type Resolver struct {
	*daeConfig.Global
}`)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)
		name, ok := structField.Tag.Lookup("mapstructure")
		if !ok {
			return "", fmt.Errorf("field %v has no required mapstructure", structField.Name)
		}
		switch field := field.Interface().(type) {
		case uint, uint8, uint16, uint32, uint64,
			int, int8, int16, int32, int64:
			// Int.
			switch field.(type) {
			case uint, uint32, uint64, int64:
				logrus.WithFields(logrus.Fields{
					"name": structField.Name,
					"type": structField.Type.String(),
				}).Debugln("converting to graphQL int32: may exceed range for large values")
			}

			b.WriteFunc(structField.Name, name, "int32", true)
		case string, bool, []string:
			b.WriteFunc(structField.Name, name, structField.Type.String(), false)
		case time.Duration:
			b.WriteMarshalFunc(structField.Name, name, "scalar.Duration")
		default:
			return "", fmt.Errorf("unknown type: %T", field)
		}
	}
	return b.sb.String(), nil
}

func main() {
	var b builder
	str, err := b.Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err = os.WriteFile(os.Args[1], []byte(str), 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
