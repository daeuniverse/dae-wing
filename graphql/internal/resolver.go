/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package internal

import (
	"github.com/v2rayA/dae/config"
	"github.com/v2rayA/dae/pkg/config_parser"
)

type FunctionOrPlaintextResolver struct {
	config.FunctionOrString
}

func (r *FunctionOrPlaintextResolver) ToPlaintext() (*PlaintextResolver, bool) {
	fs, ok := r.FunctionOrString.(string)
	if !ok {
		return nil, false
	}
	return &PlaintextResolver{Plaintext: fs}, true
}

func (r *FunctionOrPlaintextResolver) ToFunction() (*FunctionResolver, bool) {
	f, ok := r.FunctionOrString.(*config_parser.Function)
	if !ok {
		return nil, false
	}
	return &FunctionResolver{Function: f}, true
}

type AndFunctionsOrPlaintextResolver struct {
	config.FunctionListOrString
}

func (r *AndFunctionsOrPlaintextResolver) ToPlaintext() (*PlaintextResolver, bool) {
	fs, ok := r.FunctionListOrString.(string)
	if !ok {
		return nil, false
	}
	return &PlaintextResolver{Plaintext: fs}, true
}

func (r *AndFunctionsOrPlaintextResolver) ToAndFunctions() (*AndFunctionsResolver, bool) {
	fs, ok := r.FunctionListOrString.([]*config_parser.Function)
	if !ok {
		return nil, false
	}
	return &AndFunctionsResolver{AndFunctions: fs}, true
}

type PlaintextResolver struct {
	Plaintext string
}

func (r *PlaintextResolver) Val() string {
	return r.Plaintext
}

type AndFunctionsResolver struct {
	AndFunctions []*config_parser.Function
}

func (r *AndFunctionsResolver) And() (rs []*FunctionResolver) {
	for _, _f := range r.AndFunctions {
		f := _f
		rs = append(rs, &FunctionResolver{
			Function: f,
		})
	}
	return rs
}

type FunctionResolver struct {
	Function *config_parser.Function
}

func (r *FunctionResolver) Name() string {
	return r.Function.Name
}
func (r *FunctionResolver) Not() bool {
	return r.Function.Not
}
func (r *FunctionResolver) Params() (rs []*ParamResolver) {
	for _, p := range r.Function.Params {
		rs = append(rs, &ParamResolver{Param: p})
	}
	return rs
}

type ParamResolver struct {
	Param *config_parser.Param
}

func (r *ParamResolver) Key() string {
	return r.Param.Key
}
func (r *ParamResolver) Val() string {
	return r.Param.Val
}
