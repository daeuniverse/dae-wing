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

type Int64 struct {
	Int64 int64
}

// ImplementsGraphQLType maps this custom Go type
// to the graphql scalar type in the schema.
func (Int64) ImplementsGraphQLType(name string) bool {
	return name == "Int64"
}

// UnmarshalGraphQL is a custom unmarshaler for int64
//
// This function will be called whenever you use the
// Int64 scalar as an input
func (t *Int64) UnmarshalGraphQL(input interface{}) (err error) {
	switch input := input.(type) {
	case int64:
		t.Int64 = input
		return nil
	case string:
		t.Int64, err = strconv.ParseInt(input, 10, 64)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("wrong type for Int64: %v (%T)", input, input)
	}
	return nil
}

// MarshalJSON is a custom marshaler for Time
//
// This function will be called whenever you
// query for fields that use the Time type
func (t Int64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(t.Int64, 10))
}
