package node

import (
	"context"
	"testing"
	"time"

	"github.com/daeuniverse/dae-wing/common"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/graph-gophers/graphql-go"
)

func resetLatencyCache(t *testing.T) {
	t.Helper()
	nodeLatencyCache.mu.Lock()
	old := nodeLatencyCache.items
	nodeLatencyCache.items = make(map[uint]latencyCacheEntry)
	nodeLatencyCache.mu.Unlock()
	t.Cleanup(func() {
		nodeLatencyCache.mu.Lock()
		nodeLatencyCache.items = old
		nodeLatencyCache.mu.Unlock()
	})
}

func TestLatencyFreshnessPerNode(t *testing.T) {
	resetLatencyCache(t)
	now := time.Now()
	nodeLatencyCache.items[1] = latencyCacheEntry{result: &LatencyResolver{NodeID: 1}, updatedAt: now.Add(-time.Minute)}
	storeLatencyResults([]*LatencyResolver{{NodeID: 2}})
	nodes := []db.Node{{ID: 1}, {ID: 2}, {ID: 3}}
	stale := staleLatencyNodes(nodes, time.Minute, now)
	if len(stale) != 2 || stale[0].ID != 1 || stale[1].ID != 3 {
		t.Fatalf("expired and missing nodes must remain stale: %+v", stale)
	}
}

func TestQueryLatenciesScopeAndMissingConfig(t *testing.T) {
	resetLatencyCache(t)
	if err := db.InitDatabase(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	d := db.DB(ctx)
	nodes := []db.Node{{Link: "invalid-one"}, {Link: "invalid-two"}, {Link: "invalid-three"}}
	if err := d.Create(&nodes).Error; err != nil {
		t.Fatal(err)
	}
	storeLatencyResults([]*LatencyResolver{{NodeID: nodes[0].ID, AliveVal: true}})
	// Without a selected config nothing can be probed: the cached entry is
	// returned and the unprobed nodes are omitted rather than reported dead.
	results, err := QueryLatencies(ctx, nil)
	if err != nil || len(results) != 1 || results[0].NodeID != nodes[0].ID || !results[0].AliveVal {
		t.Fatalf("omitted IDs without config: results=%v err=%v", results, err)
	}
	if err := d.Create(&db.Config{Global: "global {}", Selected: true}).Error; err != nil {
		t.Fatal(err)
	}
	ids := []graphql.ID{common.EncodeCursor(nodes[1].ID)}
	results, err = QueryLatencies(ctx, &ids)
	if err != nil || len(results) != 1 || results[0].NodeID != nodes[1].ID || results[0].TestedAtV.IsZero() {
		t.Fatalf("scoped probe: results=%v err=%v", results, err)
	}
	cached := snapshotCachedLatencyResults()
	if _, ok := cached[nodes[2].ID]; ok {
		t.Fatal("scoped query probed an unrequested node")
	}
	if !cached[nodes[0].ID].AliveVal {
		t.Fatal("scoped query replaced a fresh unrelated result")
	}
	results, err = QueryLatencies(ctx, nil)
	if err != nil || len(results) != 3 || results[2].TestedAtV.IsZero() {
		t.Fatalf("omitted IDs must probe missing nodes: results=%v err=%v", results, err)
	}
	empty := []graphql.ID{}
	results, err = QueryLatencies(ctx, &empty)
	if err != nil || len(results) != 0 {
		t.Fatalf("explicit empty IDs: results=%v err=%v", results, err)
	}
}
