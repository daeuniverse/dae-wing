/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, v2rayA Organization <team@v2raya.org>
 */

package service

import "github.com/graph-gophers/graphql-go"

type PageInfoResolver struct {
	FStartCursor *graphql.ID
	FEndCursor   *graphql.ID
	FHasNextPage bool
}

func (r *PageInfoResolver) StartCursor() *graphql.ID {
	return r.FStartCursor
}

func (r *PageInfoResolver) EndCursor() *graphql.ID {
	return r.FEndCursor
}

func (r *PageInfoResolver) HasNextPage() bool {
	return r.FHasNextPage
}
