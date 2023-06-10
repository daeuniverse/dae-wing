//go:build embedallowed

/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package webrender

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/vearutop/statigz"
)

//go:embed web
var webRoot embed.FS

func Handle(mux *http.ServeMux) error {
	webFS, err := fs.Sub(webRoot, "web")
	if err != nil {
		return fmt.Errorf("fs.Sub: %w", err)
	}
	mux.Handle("/", statigz.FileServer(webFS.(fs.ReadDirFS)))
	return nil
}
