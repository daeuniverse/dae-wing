/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package node

import (
	"time"

	"github.com/daeuniverse/dae-wing/common"
	"github.com/graph-gophers/graphql-go"
)

type LatencyResolver struct {
	NodeID    uint
	LatencyMsV *int32
	AliveVal  bool
	TestedAtV time.Time
	MessageV  *string
}

func (r *LatencyResolver) ID() graphql.ID {
	return common.EncodeCursor(r.NodeID)
}

func (r *LatencyResolver) LatencyMs() *int32 {
	return r.LatencyMsV
}

func (r *LatencyResolver) Alive() bool {
	return r.AliveVal
}

func (r *LatencyResolver) TestedAt() graphql.Time {
	return graphql.Time{Time: r.TestedAtV}
}

func (r *LatencyResolver) Message() *string {
	return r.MessageV
}
