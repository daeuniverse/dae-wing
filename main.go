/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/v2rayA/dae-wing/graphql"
	"net/http"

	"github.com/graph-gophers/graphql-go/relay"
)

func main() {
	//log.Fatalln(global.SubSchema())
	schema, err := graphql.Schema()
	if err != nil {
		return
	}
	http.Handle("/query", &relay.Handler{Schema: schema})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
