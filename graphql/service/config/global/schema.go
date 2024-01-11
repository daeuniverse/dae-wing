/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package global

import (
	"fmt"
	daeConfig "github.com/daeuniverse/dae/config"
	"github.com/sirupsen/logrus"
	"github.com/stoewer/go-strcase"
	"reflect"
	"strings"
	"time"
)

type builder struct {
	sb            strings.Builder
	Head          string
	NotNullString string
}

func (b *builder) WriteLine(depth int, line string) {
	b.sb.WriteString(strings.Repeat("\t", depth))
	b.sb.WriteString(line)
	b.sb.WriteString("\n")
}

func (b *builder) Build() (string, error) {

	v := reflect.ValueOf(daeConfig.Global{})
	t := v.Type()
	b.WriteLine(0, b.Head+" {")
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)
		name, ok := structField.Tag.Lookup("mapstructure")
		if !ok {
			return "", fmt.Errorf("field %v has no required mapstructure", structField.Name)
		}
		// To lower camel case.
		name = strcase.LowerCamelCase(name)
		switch field := field.Interface().(type) {
		case uint, uint8, uint16, uint32, uint64,
			int, int8, int16, int32, int64:
			// Int.
			switch field.(type) {
			case uint, uint32, uint64, int64:
				logrus.WithFields(logrus.Fields{
					"name": structField.Name,
					"type": structField.Type.String(),
				}).Warnln("dangerous converting: may exceeds graphQL int32 range")
			}

			b.WriteLine(1, name+": Int"+b.NotNullString)
		case string:
			b.WriteLine(1, name+": String"+b.NotNullString)
		case time.Duration:
			b.WriteLine(1, name+": Duration"+b.NotNullString)
		case bool:
			b.WriteLine(1, name+": Boolean"+b.NotNullString)
		case []string:
			b.WriteLine(1, name+": [String!]"+b.NotNullString)
		default:
			return "", fmt.Errorf("unknown type: %T", field)
		}
	}
	b.WriteLine(0, "}")
	return b.sb.String(), nil
}

func Schema() (string, error) {
	typeBuilder := builder{
		sb:            strings.Builder{},
		Head:          "type Global",
		NotNullString: "!",
	}
	t, err := typeBuilder.Build()
	if err != nil {
		return "", err
	}
	inputBuilder := builder{
		sb:            strings.Builder{},
		Head:          "input globalInput",
		NotNullString: "",
	}
	i, err := inputBuilder.Build()
	if err != nil {
		return "", err
	}
	return t + "\n" + i, nil
}
