/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package node

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/dae"
	"github.com/daeuniverse/dae-wing/db"
	dialer "github.com/daeuniverse/dae/component/outbound/dialer"
	"github.com/graph-gophers/graphql-go"
	"github.com/sirupsen/logrus"
)

const latencyProbeConcurrency = 8

func TestLatencies(ctx context.Context, ids *[]graphql.ID) ([]*LatencyResolver, error) {
	option, err := latencyProbeOption(ctx)
	if err != nil {
		return nil, err
	}

	nodes, err := latencyProbeNodes(ctx, ids)
	if err != nil {
		return nil, err
	}

	results := testLatencyResultsForNodes(option, nodes)
	storeLatencyResults(results)
	return results, nil
}

func testLatencyResultsForNodes(option *dialer.GlobalOption, nodes []db.Node) []*LatencyResolver {
	results := make([]*LatencyResolver, len(nodes))
	sem := make(chan struct{}, latencyProbeConcurrency)
	var wg sync.WaitGroup

	for index := range nodes {
		index := index
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			node := nodes[index]
			results[index] = testSingleNodeLatency(option, &node)
		}()
	}

	wg.Wait()
	return results
}

func latencyProbeOption(ctx context.Context) (*dialer.GlobalOption, error) {
	var configModel db.Config
	q := db.DB(ctx).Where("selected = ?", true).First(&configModel)
	if q.Error != nil {
		return nil, q.Error
	}

	parsedConfig, err := dae.ParseConfig(&configModel.Global, nil, nil)
	if err != nil {
		return nil, err
	}

	log := logrus.New()
	log.SetOutput(io.Discard)
	return dialer.NewGlobalOption(&parsedConfig.Global, log), nil
}

func latencyProbeNodes(ctx context.Context, ids *[]graphql.ID) ([]db.Node, error) {
	q := db.DB(ctx).Model(&db.Node{})
	if ids != nil {
		decodedIDs, err := common.DecodeCursorBatch(*ids)
		if err != nil {
			return nil, err
		}
		q = q.Where("id in ?", decodedIDs)
	}

	var nodes []db.Node
	if err := q.Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func testSingleNodeLatency(option *dialer.GlobalOption, node *db.Node) *LatencyResolver {
	resolver := &LatencyResolver{
		NodeID:    node.ID,
		AliveVal:  false,
		TestedAtV: time.Now(),
	}

	d, err := dialer.NewFromLink(option, dialer.InstanceOption{DisableCheck: false}, node.Link, "")
	if err != nil {
		msg := err.Error()
		resolver.MessageV = &msg
		return resolver
	}
	defer d.Close()

	result, err := d.ProbeLatency()
	if err != nil {
		msg := err.Error()
		resolver.MessageV = &msg
		return resolver
	}

	resolver.AliveVal = result.Alive
	resolver.TestedAtV = result.CheckedAt
	if result.Alive {
		latencyMs := int32(result.Latency.Milliseconds())
		resolver.LatencyMsV = &latencyMs
		return resolver
	}
	if result.Message != "" {
		msg := result.Message
		resolver.MessageV = &msg
	}
	return resolver
}
