/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package global

import (
	_ "golang.org/x/tools/imports"
)

//go:generate go run ./generator/resolver generated_resolver.go
//go:generate go run ./generator/input generated_input.go
//go:generate go run -mod=mod golang.org/x/tools/cmd/goimports -w generated_resolver.go generated_input.go
//go:generate go fmt
