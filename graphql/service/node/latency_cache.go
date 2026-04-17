/*
 * SPDX-License-Identifier: AGPL-3.0-only
 * Copyright (c) 2023, daeuniverse Organization <team@v2raya.org>
 */

package node

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/daeuniverse/dae-wing/dae"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/graph-gophers/graphql-go"
)

var nodeLatencyCache = struct {
	mu          sync.RWMutex
	refreshMu   sync.Mutex
	updatedAt   time.Time
	items       map[uint]*LatencyResolver
}{
	items: map[uint]*LatencyResolver{},
}

func cloneLatencyResolver(resolver *LatencyResolver) *LatencyResolver {
	if resolver == nil {
		return nil
	}

	clone := *resolver
	if resolver.LatencyMsV != nil {
		latency := *resolver.LatencyMsV
		clone.LatencyMsV = &latency
	}
	if resolver.MessageV != nil {
		message := *resolver.MessageV
		clone.MessageV = &message
	}
	return &clone
}

func storeLatencyResults(results []*LatencyResolver) {
	nodeLatencyCache.mu.Lock()
	defer nodeLatencyCache.mu.Unlock()

	nodeLatencyCache.updatedAt = time.Now()
	for _, result := range results {
		if result == nil {
			continue
		}
		nodeLatencyCache.items[result.NodeID] = cloneLatencyResolver(result)
	}
}

func snapshotCachedLatencyResults() map[uint]*LatencyResolver {
	nodeLatencyCache.mu.RLock()
	defer nodeLatencyCache.mu.RUnlock()

	results := make(map[uint]*LatencyResolver, len(nodeLatencyCache.items))
	for id, resolver := range nodeLatencyCache.items {
		results[id] = cloneLatencyResolver(resolver)
	}
	return results
}

func lastLatencyCacheUpdatedAt() time.Time {
	nodeLatencyCache.mu.RLock()
	defer nodeLatencyCache.mu.RUnlock()
	return nodeLatencyCache.updatedAt
}

func loadRuntimeLatencyResults(ctx context.Context) (map[uint]*LatencyResolver, error) {
	ctl, err := dae.ControlPlane()
	if err != nil {
		if errors.Is(err, dae.ErrControlPlaneNotInit) {
			return map[uint]*LatencyResolver{}, nil
		}
		return nil, err
	}

	snapshots := ctl.SnapshotNodeLatencies()
	if len(snapshots) == 0 {
		return map[uint]*LatencyResolver{}, nil
	}

	links := make([]string, 0, len(snapshots))
	for _, snapshot := range snapshots {
		if snapshot.Link == "" {
			continue
		}
		links = append(links, snapshot.Link)
	}

	if len(links) == 0 {
		return map[uint]*LatencyResolver{}, nil
	}

	var nodes []db.Node
	if err := db.DB(ctx).Where("link in ?", links).Find(&nodes).Error; err != nil {
		return nil, err
	}

	nodeByLink := make(map[string]db.Node, len(nodes))
	for _, node := range nodes {
		nodeByLink[node.Link] = node
	}

	results := make(map[uint]*LatencyResolver)
	for _, snapshot := range snapshots {
		if snapshot.CheckedAt.IsZero() && snapshot.LatencyMs == nil && snapshot.Message == "no latency result" {
			continue
		}

		node, ok := nodeByLink[snapshot.Link]
		if !ok {
			continue
		}

		results[node.ID] = cloneLatencyResolver(&LatencyResolver{
			NodeID:     node.ID,
			LatencyMsV: snapshot.LatencyMs,
			AliveVal:   snapshot.Alive,
			TestedAtV:  snapshot.CheckedAt,
			MessageV:   stringPtr(snapshot.Message),
		})
	}

	storeLatencyResults(mapsValues(results))
	return results, nil
}

func selectedCheckInterval(ctx context.Context) (time.Duration, error) {
	var configModel db.Config
	if err := db.DB(ctx).Where("selected = ?", true).First(&configModel).Error; err != nil {
		return 0, err
	}

	parsedConfig, err := dae.ParseConfig(&configModel.Global, nil, nil)
	if err != nil {
		return 0, err
	}

	if parsedConfig.Global.CheckInterval <= 0 {
		return 30 * time.Second, nil
	}

	return parsedConfig.Global.CheckInterval, nil
}

func refreshLatencyCache(ctx context.Context) error {
	option, err := latencyProbeOption(ctx)
	if err != nil {
		return err
	}

	nodes, err := latencyProbeNodes(ctx, nil)
	if err != nil {
		return err
	}

	results := testLatencyResultsForNodes(option, nodes)
	storeLatencyResults(results)

	if ctl, err := dae.ControlPlane(); err == nil {
		ctl.TriggerLatencyChecks()
	}

	return nil
}

func refreshLatencyCacheIfNeeded(ctx context.Context) error {
	interval, err := selectedCheckInterval(ctx)
	if err != nil {
		return err
	}

	lastUpdated := lastLatencyCacheUpdatedAt()
	if !lastUpdated.IsZero() && time.Since(lastUpdated) < interval {
		return nil
	}

	nodeLatencyCache.refreshMu.Lock()
	defer nodeLatencyCache.refreshMu.Unlock()

	lastUpdated = lastLatencyCacheUpdatedAt()
	if !lastUpdated.IsZero() && time.Since(lastUpdated) < interval {
		return nil
	}

	return refreshLatencyCache(ctx)
}

func mapsValues(items map[uint]*LatencyResolver) []*LatencyResolver {
	results := make([]*LatencyResolver, 0, len(items))
	for _, item := range items {
		results = append(results, item)
	}
	return results
}

func stringPtr(value string) *string {
	return &value
}

func QueryLatencies(ctx context.Context, ids *[]graphql.ID) ([]*LatencyResolver, error) {
	_ = refreshLatencyCacheIfNeeded(ctx)

	nodes, err := latencyProbeNodes(ctx, ids)
	if err != nil {
		return nil, err
	}

	merged := snapshotCachedLatencyResults()
	runtimeResults, err := loadRuntimeLatencyResults(ctx)
	if err != nil {
		return nil, err
	}
	for id, resolver := range runtimeResults {
		merged[id] = resolver
	}

	results := make([]*LatencyResolver, 0, len(nodes))
	for _, node := range nodes {
		if resolver, ok := merged[node.ID]; ok {
			results = append(results, cloneLatencyResolver(resolver))
		}
	}

	return results, nil
}
