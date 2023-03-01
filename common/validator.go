/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package common

import "regexp"

func ValidateRemarks(remarks string) bool {
	return regexp.MustCompile(`^[^:']+$`).MatchString(remarks)
}

func ValidateId(id string) bool {
	// https://github.com/v2rayA/dae-config-dist/blob/main/dae_config.g4
	return regexp.MustCompile(`^[a-zA-Z_][-a-zA-Z0-9_/\\^*+.=@$!#%]*$`).MatchString(id)
}
