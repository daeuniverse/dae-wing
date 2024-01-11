/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package scalar

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Uint64 struct {
	Uint64 uint64
}

// ImplementsGraphQLType maps this custom Go type
// to the graphql scalar type in the schema.
func (Uint64) ImplementsGraphQLType(name string) bool {
	return name == "Uint64"
}

// UnmarshalGraphQL is a custom unmarshaler for uint64
//
// This function will be called whenever you use the
// Uint64 scalar as an input
func (t *Uint64) UnmarshalGraphQL(input interface{}) (err error) {
	switch input := input.(type) {
	case uint64:
		t.Uint64 = input
		return nil
	case string:
		t.Uint64, err = strconv.ParseUint(input, 10, 64)
		if err != nil {
			return err
		}
	case float64:
		if uint64(input*10)%10 != 0 {
			return fmt.Errorf("wrong type for Uint64: %v (%T)", input, input)
		}
		t.Uint64 = uint64(input)
	default:
		return fmt.Errorf("wrong type for Uint64: %v (%T)", input, input)
	}
	return nil
}

// MarshalJSON is a custom marshaler for Time
//
// This function will be called whenever you
// query for fields that use the Time type
func (t Uint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(t.Uint64, 10))
}
