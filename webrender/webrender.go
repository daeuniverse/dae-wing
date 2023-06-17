//go:build embedallowed

/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package webrender

import (
	"compress/gzip"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/vearutop/statigz"
)

//go:embed web
var webRoot embed.FS

func Handle(mux *http.ServeMux) error {
	webFS, err := fs.Sub(webRoot, "web")
	if err != nil {
		return fmt.Errorf("fs.Sub: %w", err)
	}
	mux.Handle("/", statigz.FileServer(webFS.(fs.ReadDirFS), statigz.OnNotFound(func(rw http.ResponseWriter, req *http.Request) {
		// Fallback to index.html.
		indexHtml, err := webFS.Open("index.html")
		indexHtmlGziped := false
		var r io.Reader
		if err != nil {
			if os.IsNotExist(err) {
				indexHtml, err = webFS.Open("index.html.gz")
				if err != nil {
					goto errNoIndexHtml
				}
				indexHtmlGziped = true
			} else {
				goto errNoIndexHtml
			}
		}
		defer indexHtml.Close()
		r = indexHtml
		if indexHtmlGziped {
			acceptGzip := false
			for _, e := range strings.Split(req.Header.Get("Accept-Encoding"), ",") {
				if strings.TrimSpace(e) == "gzip" {
					acceptGzip = true
					break
				}
			}
			if acceptGzip {
				rw.Header().Set("Content-Encoding", "gzip")
			} else {
				var err error
				r, err = gzip.NewReader(indexHtml)
				if err != nil {
					rw.WriteHeader(400)
					return
				}
			}
		}
		_, _ = io.Copy(rw, r)
		return
	errNoIndexHtml:
		rw.WriteHeader(404)
		return

	})))
	return nil
}
