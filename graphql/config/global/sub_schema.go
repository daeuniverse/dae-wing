/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package global

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/v2rayA/dae/config"
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

func (b *builder) Build() (string, error) {

	v := reflect.ValueOf(config.Global{})
	t := v.Type()
	b.WriteLine(0, "type Global {")
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
				}).Warnln("dangerous converting: may exceeds graphQL int32 range")
			}

			b.WriteLine(1, name+": Int!")
		case string:
			b.WriteLine(1, name+": String!")
		case time.Duration:
			b.WriteLine(1, name+": Duration!")
		case bool:
			b.WriteLine(1, name+": Boolean!")
		case []string:
			b.WriteLine(1, name+": [String!]!")
		default:
			return "", fmt.Errorf("unknown type: %T", field)
		}
	}
	b.WriteLine(0, "}")
	return b.sb.String(), nil
}

func SubSchema() (string, error) {
	var b builder
	return b.Build()
}
