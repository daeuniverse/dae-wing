/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package scalar

import (
	"encoding/json"
	"fmt"
	"time"
)

type Duration struct {
	time.Duration
}

// ImplementsGraphQLType maps this custom Go type
// to the graphql scalar type in the schema.
func (Duration) ImplementsGraphQLType(name string) bool {
	return name == "Duration"
}

// UnmarshalGraphQL is a custom unmarshaler for Time.Duration
//
// This function will be called whenever you use the
// Duration scalar as an input
func (t *Duration) UnmarshalGraphQL(input interface{}) (err error) {
	switch input := input.(type) {
	case time.Duration:
		t.Duration = input
		return nil
	case string:
		t.Duration, err = time.ParseDuration(input)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("wrong type for Time: %T", input)
	}
	return nil
}

// MarshalJSON is a custom marshaler for Time
//
// This function will be called whenever you
// query for fields that use the Time type
func (t Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Duration.String())
}
