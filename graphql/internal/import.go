/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package internal

import "github.com/v2rayA/dae-wing/common"

type ImportArgument struct {
	Link string
	Tag  *string
}

func (a *ImportArgument) ValidateTag() error {
	if a.Tag == nil {
		return nil
	}
	return common.ValidateTag(*a.Tag)
}
