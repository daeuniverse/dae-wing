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
	"gorm.io/gorm"
)

type latencyCacheEntry struct {
	result    *LatencyResolver
	updatedAt time.Time
}

var nodeLatencyCache = struct {
	mu        sync.RWMutex
	refreshMu sync.Mutex
	items     map[uint]latencyCacheEntry
}{
	items: map[uint]latencyCacheEntry{},
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

	now := time.Now()
	for _, result := range results {
		if result == nil {
			continue
		}
		nodeLatencyCache.items[result.NodeID] = latencyCacheEntry{
			result: cloneLatencyResolver(result), updatedAt: now,
		}
	}
}

func snapshotCachedLatencyResults() map[uint]*LatencyResolver {
	nodeLatencyCache.mu.RLock()
	defer nodeLatencyCache.mu.RUnlock()

	results := make(map[uint]*LatencyResolver, len(nodeLatencyCache.items))
	for id, resolver := range nodeLatencyCache.items {
		results[id] = cloneLatencyResolver(resolver.result)
	}
	return results
}

func staleLatencyNodes(nodes []db.Node, interval time.Duration, now time.Time) []db.Node {
	nodeLatencyCache.mu.RLock()
	defer nodeLatencyCache.mu.RUnlock()
	var stale []db.Node
	for _, node := range nodes {
		entry, ok := nodeLatencyCache.items[node.ID]
		if !ok || now.Sub(entry.updatedAt) >= interval {
			stale = append(stale, node)
		}
	}
	return stale
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

	nodesByLink := make(map[string][]db.Node, len(nodes))
	for _, node := range nodes {
		nodesByLink[node.Link] = append(nodesByLink[node.Link], node)
	}

	results := make(map[uint]*LatencyResolver)
	for _, snapshot := range snapshots {
		if snapshot.CheckedAt.IsZero() && snapshot.LatencyMs == nil && snapshot.Message == "no latency result" {
			continue
		}

		matchedNodes, ok := nodesByLink[snapshot.Link]
		if !ok {
			continue
		}

		for _, node := range matchedNodes {
			results[node.ID] = cloneLatencyResolver(&LatencyResolver{
				NodeID:     node.ID,
				LatencyMsV: snapshot.LatencyMs,
				AliveVal:   snapshot.Alive,
				TestedAtV:  snapshot.CheckedAt,
				MessageV:   stringPtr(snapshot.Message),
			})
		}
	}

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

func refreshLatencyCache(ctx context.Context, nodes []db.Node, all bool) error {
	option, err := latencyProbeOption(ctx)
	if err != nil {
		return err
	}

	results := testLatencyResultsForNodes(option, nodes)
	storeLatencyResults(results)

	if all {
		if ctl, err := dae.ControlPlane(); err == nil {
			ctl.TriggerLatencyChecks()
		}
	}

	return nil
}

func refreshLatencyCacheIfNeeded(ctx context.Context, nodes []db.Node, all bool) error {
	interval, err := selectedCheckInterval(ctx)
	if err != nil {
		return err
	}

	if len(staleLatencyNodes(nodes, interval, time.Now())) == 0 {
		return nil
	}

	nodeLatencyCache.refreshMu.Lock()
	defer nodeLatencyCache.refreshMu.Unlock()

	stale := staleLatencyNodes(nodes, interval, time.Now())
	if len(stale) == 0 {
		return nil
	}

	return refreshLatencyCache(ctx, stale, all)
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
	nodes, err := latencyProbeNodes(ctx, ids)
	if err != nil {
		return nil, err
	}
	if err := refreshLatencyCacheIfNeeded(ctx, nodes, ids == nil); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	merged := snapshotCachedLatencyResults()
	runtimeResults, err := loadRuntimeLatencyResults(ctx)
	if err != nil {
		return nil, err
	}
	for id, resolver := range runtimeResults {
		if _, ok := merged[id]; !ok {
			merged[id] = resolver
		}
	}

	results := make([]*LatencyResolver, 0, len(nodes))
	for _, node := range nodes {
		if resolver, ok := merged[node.ID]; ok {
			results = append(results, cloneLatencyResolver(resolver))
		}
	}

	return results, nil
}
