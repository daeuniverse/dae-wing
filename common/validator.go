/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package common

import (
	"fmt"
	"regexp"
)

var InvalidIdFormatError = fmt.Errorf("invalid id; only support numbers and letters")
var InvalidTagFormatError = fmt.Errorf("invalid tag; cannot contains `:` or `'`")

func ValidateTag(tag string) error {
	if !regexp.MustCompile(`^[^:']+$`).MatchString(tag){
		return InvalidTagFormatError
	}
	return nil
}

func ValidateId(id string) error {
	// https://github.com/v2rayA/dae-config-dist/blob/main/dae_config.g4
	if  !regexp.MustCompile(`^[a-zA-Z_][-a-zA-Z0-9_/\\^*+.=@$!#%]*$`).MatchString(id){
		return InvalidIdFormatError
	}
	return nil
}
