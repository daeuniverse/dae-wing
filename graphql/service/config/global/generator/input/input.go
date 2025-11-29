/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	daeConfig "github.com/daeuniverse/dae/config"
	"github.com/sirupsen/logrus"
	"github.com/stoewer/go-strcase"
)

type builder struct {
	// Struct builder.
	st strings.Builder
	// Assign method builder.
	me strings.Builder
}

func (b *builder) WriteLine(depth int, line string) {
	b.st.WriteString(strings.Repeat("\t", depth))
	b.st.WriteString(line)
	b.st.WriteString("\n")
}

func (b *builder) WriteMethodLine(depth int, line string) {
	b.me.WriteString(strings.Repeat("\t", depth))
	b.me.WriteString(line)
	b.me.WriteString("\n")
}

func (b *builder) WriteField(name string, retTyp string) {
	b.WriteLine(1, fmt.Sprintf("%v\t*%v", strcase.UpperCamelCase(name), retTyp))
}

func (b *builder) WriteMethodTransform(fieldName string, name string, retTyp string, cast bool) {
	b.WriteMethodLine(1, fmt.Sprintf("if i.%v != nil {", strcase.UpperCamelCase(name)))
	var right string
	if cast {
		right = fmt.Sprintf("%v(*i.%v)", retTyp, strcase.UpperCamelCase(name))
	} else {
		right = fmt.Sprintf("*i.%v", strcase.UpperCamelCase(name))
	}
	b.WriteMethodLine(2, fmt.Sprintf("g.%v = %v", fieldName, right))
	b.WriteMethodLine(1, fmt.Sprintf("}"))
}
func (b *builder) WriteMethodScalar(fieldName string, name string, scalarField string) {
	b.WriteMethodLine(1, fmt.Sprintf("if i.%v != nil {", strcase.UpperCamelCase(name)))
	b.WriteMethodLine(2, fmt.Sprintf("g.%v = i.%v.%v", fieldName, strcase.UpperCamelCase(name), scalarField))
	b.WriteMethodLine(1, fmt.Sprintf("}"))
}

func (b *builder) Build() (string, error) {

	v := reflect.ValueOf(daeConfig.Global{})
	t := v.Type()
	b.WriteLine(0, "// Generated code; DO NOT EDIT.\n")
	b.WriteLine(0, "package global\n")
	b.WriteLine(0, fmt.Sprintf(`import "%v"`, t.PkgPath()))
	b.WriteLine(0, fmt.Sprintf(`import "github.com/daeuniverse/dae-wing/graphql/scalar"`))
	b.WriteLine(0, fmt.Sprintf(`import daeConfig "github.com/daeuniverse/dae/config"`))
	b.WriteLine(0, `type Input struct {`)
	b.WriteMethodLine(0, `func (i *Input) Assign(g *daeConfig.Global) {`)
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

			b.WriteField(name, "int32")
			b.WriteMethodTransform(structField.Name, name, structField.Type.String(), true)
		case string, bool, []string:
			b.WriteField(name, structField.Type.String())
			b.WriteMethodTransform(structField.Name, name, structField.Type.String(), false)
		case time.Duration:
			b.WriteField(name, "scalar.Duration")
			b.WriteMethodScalar(structField.Name, name, "Duration")
		default:
			return "", fmt.Errorf("unknown type: %T", field)
		}
	}
	b.WriteLine(0, `}`)
	b.WriteMethodLine(0, `}`)
	b.WriteMethodLine(0, `
func (i *Input) Marshal() (string, error) {
			var g daeConfig.Global
			i.Assign(&g)
			marshaller := daeConfig.Marshaller{
				IgnoreZero: true,
			}
			if err := marshaller.MarshalSection("global", reflect.ValueOf(g), 0); err != nil {
				return "", err
			}
			return string(marshaller.Bytes()), nil
}`)
	return b.st.String() + "\n" + b.me.String(), nil
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
