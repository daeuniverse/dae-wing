/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package global

import (
	_ "golang.org/x/tools/imports"
)

//go:generate go run ./generator generated_resolver.go
//go:generate go run golang.org/x/tools/cmd/goimports -w generated_resolver.go
//go:generate go fmt
