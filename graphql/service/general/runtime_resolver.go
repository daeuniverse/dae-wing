/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package general

import (
	"strconv"

	"github.com/daeuniverse/dae-wing/dae"
	"github.com/graph-gophers/graphql-go"
)

type RuntimeOverviewResolver struct {
	Overview *dae.RuntimeOverview
}

type RuntimeTrafficSampleResolver struct {
	Sample dae.RuntimeTrafficSample
}

func (r *Resolver) RuntimeOverview(args *struct {
	WindowSec int32
	MaxPoints int32
}) (*RuntimeOverviewResolver, error) {
	overview, err := dae.GetRuntimeOverview(int(args.WindowSec), int(args.MaxPoints))
	if err != nil {
		return nil, err
	}
	return &RuntimeOverviewResolver{Overview: overview}, nil
}

func (r *RuntimeOverviewResolver) UpdatedAt() graphql.Time {
	return graphql.Time{Time: r.Overview.UpdatedAt}
}

func (r *RuntimeOverviewResolver) UploadRate() float64 {
	return float64(r.Overview.UploadRate)
}

func (r *RuntimeOverviewResolver) DownloadRate() float64 {
	return float64(r.Overview.DownloadRate)
}

func (r *RuntimeOverviewResolver) UploadTotal() string {
	return strconv.FormatUint(r.Overview.UploadTotal, 10)
}

func (r *RuntimeOverviewResolver) DownloadTotal() string {
	return strconv.FormatUint(r.Overview.DownloadTotal, 10)
}

func (r *RuntimeOverviewResolver) ActiveConnections() int32 {
	return int32(r.Overview.ActiveConnections)
}

func (r *RuntimeOverviewResolver) UdpSessions() int32 {
	return int32(r.Overview.UDPSessions)
}

func (r *RuntimeOverviewResolver) Samples() []*RuntimeTrafficSampleResolver {
	resolvers := make([]*RuntimeTrafficSampleResolver, 0, len(r.Overview.Samples))
	for _, sample := range r.Overview.Samples {
		resolvers = append(resolvers, &RuntimeTrafficSampleResolver{Sample: sample})
	}
	return resolvers
}

func (r *RuntimeTrafficSampleResolver) Timestamp() graphql.Time {
	return graphql.Time{Time: r.Sample.Timestamp}
}

func (r *RuntimeTrafficSampleResolver) UploadRate() float64 {
	return float64(r.Sample.UploadRate)
}

func (r *RuntimeTrafficSampleResolver) DownloadRate() float64 {
	return float64(r.Sample.DownloadRate)
}
