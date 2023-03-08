/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package dae

import (
	daeConfig "github.com/v2rayA/dae/config"
	"reflect"
)

type FlatDesc struct {
	Name         string `json:"name,omitempty"`
	Mapping      string `json:"mapping,omitempty"`
	IsArray      bool   `json:"isArray,omitempty"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Required     bool   `json:"required,omitempty"`
	Type         string `json:"type,omitempty"`
	Desc         string `json:"desc,omitempty"`
}

func ExportFlatDesc() []*FlatDesc {
	t := reflect.TypeOf(daeConfig.Config{})
	exporter := flatDescExporter{
		leaves:       make(map[string]reflect.Type),
		pkgPathScope: t.PkgPath(),
	}
	descList := exporter.exportStruct("", "", t, daeConfig.SectionSummaryDesc, false)
	return descList
}

type flatDescExporter struct {
	leaves       map[string]reflect.Type
	pkgPathScope string
}

func (e *flatDescExporter) exportStruct(namePrefix string, mappingPrefix string, t reflect.Type, descSource daeConfig.Desc, inheritSource bool) (descList []*FlatDesc) {
	for i := 0; i < t.NumField(); i++ {
		section := t.Field(i)
		mapping := section.Tag.Get("mapstructure")
		// Parse desc.
		var desc string
		if descSource != nil {
			desc = descSource[mapping]
		}
		// Parse elem type.
		var isArray bool
		var typ reflect.Type
		switch section.Type.Kind() {
		case reflect.Slice:
			typ = section.Type.Elem()
			isArray = true
		default:
			typ = section.Type
		}
		if typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		// Parse children.
		var children []*FlatDesc
		switch typ.Kind() {
		case reflect.Struct:
			var nextDescSource daeConfig.Desc
			if inheritSource {
				nextDescSource = descSource
			} else {
				nextDescSource = daeConfig.SectionDescription[section.Tag.Get("desc")]
			}
			if typ.PkgPath() == "" || typ.PkgPath() == e.pkgPathScope {
				children = e.exportStruct(
					namePrefix+section.Name+".",
					mappingPrefix+mapping+".",
					typ,
					nextDescSource,
					true,
				)
			}
		}
		if len(children) == 0 {
			// Record leaves.
			e.leaves[typ.String()] = typ
		}
		_, required := section.Tag.Lookup("required")
		descList = append(descList, &FlatDesc{
			Name:         namePrefix + section.Name,
			Mapping:      mappingPrefix + mapping,
			IsArray:      isArray,
			DefaultValue: section.Tag.Get("default"),
			Required:     required,
			Type:         typ.String(),
			Desc:         desc,
		})
		descList = append(descList, children...)
	}
	return descList
}
