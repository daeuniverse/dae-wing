/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package scalar

import (
	"encoding/json"
	"fmt"
)

type Int8 struct {
	int8
}

func (Int8) ImplementsGraphQLType(name string) bool {
	return name == "Int8"
}
func (t *Int8) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case int8:
		t.int8 = input
		return nil
	default:
		return fmt.Errorf("wrong type for Time: %T", input)
	}
}
func (t Int8) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.int8)
}
