//go:build !embedallowed

/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package webrender

import "net/http"

func Handle(mux *http.ServeMux) error {
	return nil
}
